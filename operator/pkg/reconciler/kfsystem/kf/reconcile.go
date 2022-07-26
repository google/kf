// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kf

import (
	"context"
	"fmt"
	"time"

	"kf-operator/pkg/apis/kfsystem/v1alpha1"
	operandv1alpha1 "kf-operator/pkg/apis/operand/v1alpha1"
	"kf-operator/pkg/operand"
	kfoperand "kf-operator/pkg/operand/kf"
	"kf-operator/pkg/operand/transformations"

	"github.com/Masterminds/semver/v3"
	mf "github.com/manifestival/manifestival"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	v1listers "k8s.io/client-go/listers/core/v1"
	certresources "knative.dev/pkg/webhook/certificates/resources"
)

const (
	versionLabel                  = "app.kubernetes.io/version"
	workloadIdentityAnnotationKey = "iam.gke.io/gcp-service-account"
	kfServiceAccountName          = "controller"
	kccProjectAnnotationKey       = "cnrm.cloud.google.com/project-id"

	subresourceAPIServerName  = "subresource-apiserver"
	subresourceAPIServiceName = "subresource-apiserver.kf.svc"
	kfControllerServerName    = "controller"
	kfNamespace               = "kf"
	oneWeek                   = 7 * 24 * time.Hour
)

// Reconciler reconciles kf.
type Reconciler struct {
	operand.Factory
	Versions     []*semver.Version
	Lookup       func(string) (*mf.Manifest, error)
	SecretLister v1listers.SecretLister
	PodLister    v1listers.PodLister
}

var _ kfoperand.Ctor = (*Reconciler)(nil)

// CalculateOperand calculates the operand this reconciliation cycle.
func (e Reconciler) CalculateOperand(ctx context.Context, cr *v1alpha1.KfSystem) (*operandv1alpha1.OperandSpec, error) {
	var t *mf.Manifest
	if !cr.Spec.Kf.IsEnabled() {
		cr.Status.TargetKfVersion = ""
		return &operandv1alpha1.OperandSpec{}, nil
	}
	if cr.Spec.Kf.Version == "" {
		cr.Spec.Kf.Version = "latest"
	}
	v := cr.Spec.Kf.Version

	if cr.Status.KfVersion != v {
		tv, err := e.findNextVersionLeadingTo(cr)
		if err != nil {
			return nil, err
		}
		cr.Status.TargetKfVersion = *tv
	}
	eval, err := e.Lookup(cr.Status.TargetKfVersion)
	if err != nil {
		return nil, err
	}
	t = eval

	var key, cert, caCert []byte
	apiServiceSecret, err := e.SecretLister.
		Secrets(kfNamespace).
		Get(apiServiceSecretName)
	if !apierrs.IsNotFound(err) && err != nil {
		return nil, fmt.Errorf("failed to get API server secret: %v", err)
	} else if apiServiceSecret != nil {
		cert = apiServiceSecret.Data[certresources.ServerCert]
		key = apiServiceSecret.Data[certresources.ServerKey]
		caCert = apiServiceSecret.Data[certresources.CACert]
	}

	if len(key) == 0 || len(cert) == 0 || len(caCert) == 0 {
		var err error
		// Bootstrap certs for the API service. The Kf controller
		// manages the certs, however this implies there is a period of
		// time where the subresource API server is waiting on proper
		// certs.
		key, cert, caCert, err = certresources.CreateCerts(
			ctx,
			subresourceAPIServiceName,
			kfNamespace,

			// Give the bootstrap certs a week of runtime. The Kf
			// controller will create new certs as these approach
			// their expiration.
			time.Now().Add(oneWeek),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create API server bootstrap certs: %v", err)
		}
	}

	transformers := []mf.Transformer{
		AddFeatureFlags(ctx, cr.Spec.Kf.Config.FeatureFlags),
		GatewayTransform(ctx, cr.Spec.Kf.Config.Gateway),
		ConfigSecretsTransform(ctx, cr.Spec.Kf.Config.Secrets),
		ConfigDefaultsTransform(ctx, cr.Spec.Kf.Config.DefaultsConfig),
		APIServerCertsTransformer(ctx, caCert),
		APIServerSecretTransformer(ctx, cert, key, caCert),
	}

	{
		// If the subresource API server doesn't have any pods scheduled for
		// it yet, then we don't want to have the API service enabled. Doing
		// so implies the API server will try to hit the missing pods and
		// could be deemed unhealthy.
		pods, err := e.PodLister.Pods(kfNamespace).List(labels.SelectorFromValidatedSet(labels.Set{
			"app": "subresource-apiserver",
		}))
		if err != nil {
			return nil, fmt.Errorf("failed listing subresource API server pods: %v", err)
		}
		if len(pods) == 0 {
			// There aren't any pods scheduled, so to be on the safe side,
			// remove the API server.
			transformers = append(transformers, APIServerRemoverTransformer(ctx))
		}
	}

	// We require the ControllerCACerts secret to be immutable so that if the
	// user wants to change out the certs, they have to create a new secret
	// and update the spec. This will force a reconcillation loop and update
	// the pods accordingly.
	if name := cr.Spec.Kf.Config.Secrets.ControllerCACerts.Name; name != "" {
		secret, err := e.SecretLister.Secrets(kfNamespace).Get(name)

		// NOTE: Any error (even NotFound) should be returned. If the user
		// provided a secret, it must exist already.
		if err != nil {
			return nil, err
		}

		if secret.Immutable == nil || !(*secret.Immutable) {
			return nil, fmt.Errorf("secret %s/%s must be immutable", secret.Namespace, secret.Name)
		}

		transformers = append(transformers, CertVolumeTransformer(ctx, secret))
	}

	secrets := cr.Spec.Kf.Config.Secrets
	if secrets.WorkloadIdentity != nil {
		wi := getGoogleServiceAccount(secrets.WorkloadIdentity.GoogleServiceAccount, secrets.WorkloadIdentity.GoogleProjectID)
		transformers = append(transformers,
			transformations.AddAnnotation(ctx, "ServiceAccount", kfServiceAccountName, map[string]string{workloadIdentityAnnotationKey: wi}),
			transformations.AddAnnotation(ctx, "Namespace", "kf", map[string]string{kccProjectAnnotationKey: secrets.WorkloadIdentity.GoogleProjectID}),
			transformations.AddWICheckForSubresourceAPI(),
		)
	} else if secrets.Build != nil && secrets.Build.ImagePushSecretName != "" {
		transformers = append(transformers, transformations.AppendDockerCredentials(subresourceAPIServerName, secrets.Build.ImagePushSecretName))
		// kf controller uses the docker credentials to read metadata (entrypoint) of App container image for running Kf Tasks.
		transformers = append(transformers, transformations.AppendDockerCredentials(kfControllerServerName, secrets.Build.ImagePushSecretName))
	}

	return e.FromGeneralManifest(
		ctx,
		true,
		*t,
		transformers...,
	)
}

// MarkInstallSucceeded marks kf install succeeded.
func (e Reconciler) MarkInstallSucceeded(kfs *v1alpha1.KfSystem) {
	kfs.Status.MarkKfInstallSucceeded(kfs.Status.TargetKfVersion)
}

// MarkInstallNotReady marks kf install not ready.
func (e Reconciler) MarkInstallNotReady(kfs *v1alpha1.KfSystem) {
	kfs.Status.MarkKfInstallNotReady()
}

// MarkInstallFailed marks kf install failed.
func (e Reconciler) MarkInstallFailed(kfs *v1alpha1.KfSystem, err string) {
	kfs.Status.MarkKfInstallFailed(err)
}

func (e Reconciler) findNextVersionLeadingTo(cr *v1alpha1.KfSystem) (*string, error) {
	v := cr.Spec.Kf.Version
	pvs := cr.Status.KfVersion
	if v == "latest" {
		v = e.Versions[len(e.Versions)-1].Original()
	}
	if pvs == "" {
		return &v, nil
	}
	pv, err := semver.NewVersion(pvs)
	if err != nil {
		return nil, fmt.Errorf("unable to read previous version: %+v", err)
	}
	dv, err := semver.NewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("unable to read desired version: %+v", err)
	}

	if pv.Major() != dv.Major() {
		return nil, fmt.Errorf("major version changes (%d -> %d) are not yet supported", pv.Major(), dv.Major())
	}

	var nm uint64
	if dv.Equal(pv) {
		return &v, nil
	}
	if dv.GreaterThan(pv) {
		nm = pv.Minor() + 1
	} else if dv.LessThan(pv) {
		nm = pv.Minor() - 1
	}

	var c *semver.Constraints
	if nm == dv.Minor() {
		c, err = semver.NewConstraint(dv.Original())
	} else {
		c, err = semver.NewConstraint(fmt.Sprintf("%d.%d", dv.Major(), nm))
	}

	if err != nil {
		return nil, err
	}

	for i := range e.Versions {
		// Reverse iteration order so that we find the latest matching version.
		if cv := e.Versions[len(e.Versions)-i-1]; c.Check(cv) {
			t := cv.Original()
			return &t, nil
		}
	}
	return nil, fmt.Errorf("Unable to find version which allows for progression to %s", dv.Original())
}
