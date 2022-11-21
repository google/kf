// Copyright 2019 Google LLC

package resources

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/internal/selectorutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
)

const (
	// DefaultRevisionHistoryLimit contains the default number of revisions to be
	// able to roll back/forward to.
	DefaultRevisionHistoryLimit int32 = 10

	// kfMounts is the where the real NFS share is mounted.
	kfMounts string = "/.kfmounts"

	// fuseMountPrefix is the prefix that's added to the name of the fuse Volume shared between containers.
	fuseMountPrefix string = "fuse"

	// buildpackV2VcapUserId is the user ID of the running process in buildpack v2 apps.
	buildpackV2VcapUserId int64 = 2000

	// mapfsCmdFormat defines the format for the mapfs command.
	mapfsCmdFormat string = "mapfs -uid %d -gid %d %s %s &"

	// unmountCmdFormat defines the format for the unmount command.
	unmountCmdFormat string = "fusermount -u %s &"
)

var (
	// defaultMaxSurge and defaultMaxUnavailable are the default values for Deployment's rolling upgrade strategy.
	defaultMaxSurge       intstr.IntOrString = intstr.FromString("25%")
	defaultMaxUnavailable intstr.IntOrString = intstr.FromString("25%")
)

// DeploymentName gets the name of a Deployment given the app.
func DeploymentName(app *v1alpha1.App) string {
	return app.Name
}

// MakeDeployment creates a K8s Deployment from an app definition.
func MakeDeployment(
	app *v1alpha1.App,
	space *v1alpha1.Space,
) (*appsv1.Deployment, error) {
	image := app.Status.Image
	if image == "" {
		return nil, errors.New("waiting for build image in latestReadyBuild")
	}

	replicas, err := app.Spec.Instances.DeploymentReplicas()
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(app, space)

	if err != nil {
		return nil, err
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName(app),
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(app),
			},
			Labels: v1alpha1.UnionMaps(app.GetLabels(), app.ComponentLabels("app-scaler")),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: metav1.SetAsLabelSelector(labels.Set(PodLabels(app))),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: v1alpha1.UnionMaps(
						app.GetLabels(),
						// Add in the App's labels, which may be user-defined.
						PodLabels(app),

						// Insert a label for isolating apps with their own NetworkPolicies.
						map[string]string{
							v1alpha1.NetworkPolicyLabel: v1alpha1.NetworkPolicyApp,
						}),
					Annotations: v1alpha1.UnionMaps(
						// Add in the App's annotations, which may be user-defined.
						app.GetAnnotations(),

						map[string]string{
							// Inject the Envoy sidecar on all apps so networking rules
							// apply.
							"sidecar.istio.io/inject":                          "true",
							"traffic.sidecar.istio.io/includeOutboundIPRanges": "*",

							// Follow KEP-2227 which allows setting the default
							// container name for kubectl logs/exec/debug/attach etc.
							"kubectl.kubernetes.io/default-container": v1alpha1.DefaultUserContainerName,
						}),
				},
				Spec: *podSpec,
			},
			RevisionHistoryLimit: ptr.Int32(DefaultRevisionHistoryLimit),
			Replicas:             ptr.Int32(int32(replicas)),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &defaultMaxUnavailable,
					MaxSurge:       &defaultMaxSurge,
				},
			},
			ProgressDeadlineSeconds: space.Status.RuntimeConfig.ProgressDeadlineSeconds,
		},
	}, nil
}

func makePodSpec(app *v1alpha1.App, space *v1alpha1.Space) (*corev1.PodSpec, error) {
	// don't modify the spec on the app
	spec := app.Spec.Template.Spec.DeepCopy()

	// Don't inject the old Docker style environment variables for every service
	// doing so could cause misconfiguration for apps.
	spec.EnableServiceLinks = ptr.Bool(false)

	// At this point in the lifecycle there should be exactly one container
	// if the webhhook is working but create one to avoid panics just in case.
	if len(spec.Containers) == 0 {
		spec.Containers = append(spec.Containers, corev1.Container{})
	}

	userPort := getUserPort(app)
	userContainer := &spec.Containers[0]
	userContainer.Name = v1alpha1.DefaultUserContainerName
	userContainer.Image = app.Status.Image
	// If the user hasn't overwritten the port, open one by default.
	if len(userContainer.Ports) == 0 {
		userContainer.Ports = buildContainerPorts(userPort)
	}

	// Execution environment variables come before others because they're built
	// to be overridden.
	containerEnv := []corev1.EnvVar{}
	containerEnv = append(containerEnv, space.Status.RuntimeConfig.Env...)
	containerEnv = append(containerEnv, userContainer.Env...)

	// Add in additinal CF style environment variables
	containerEnv = append(containerEnv, BuildRuntimeEnvVars(CFRunning, app)...)

	// XXX: Add a no-op environment variable that reflects the UpdateRequests.
	// This will force K8s to restart the pods.
	containerEnv = append(containerEnv, corev1.EnvVar{
		Name:  fmt.Sprintf("KF_UPDATE_REQUESTS_%v", app.UID),
		Value: strconv.FormatInt(int64(app.Spec.Template.UpdateRequests), 10),
	})

	userContainer.Env = containerEnv

	// Explicitly disable stdin and tty allocation
	userContainer.Stdin = false
	userContainer.TTY = false

	// Populate default container values
	userContainer.ImagePullPolicy = corev1.PullIfNotPresent
	userContainer.TerminationMessagePath = corev1.TerminationMessagePathDefault
	userContainer.TerminationMessagePolicy = corev1.TerminationMessageReadFile

	// If the client provides probes, we should fill in the port for them.
	rewriteUserProbe(userContainer.LivenessProbe, userPort)
	rewriteUserProbe(userContainer.ReadinessProbe, userPort)
	rewriteUserProbe(userContainer.StartupProbe, userPort)

	spec.ServiceAccountName = app.Status.ServiceAccountName
	// This need to be removed after we implement server side apply.
	spec.DeprecatedServiceAccount = spec.ServiceAccountName

	// If the client provides a node selector, we should fill in the corresponding field in the podspec.
	spec.NodeSelector = selectorutil.GetNodeSelector(app.Spec.Build.Spec, space)

	if len(app.Status.Volumes) > 0 {
		// mapfs for volumes needs the extra permission.
		userContainer.SecurityContext = &corev1.SecurityContext{
			Privileged: ptr.Bool(true),
		}
		// build nfs volume mounts.
		volumes, userVolumeMounts, fuseCommands, unmountCommands, err := buildVolumes(app.Status.Volumes)
		if err != nil {
			return nil, err
		}
		spec.Volumes = append(spec.Volumes, volumes...)
		userContainer.VolumeMounts = append(userContainer.VolumeMounts, userVolumeMounts...)

		originalArgs := userContainer.Args
		originalCommand := []string{}
		if len(userContainer.Command) > 0 {
			// Append to the existing array so we don't modify the userContainer.Command value.
			originalCommand = append(originalCommand, userContainer.Command...)
		} else {
			// TODO: Look up the correct value rather than assuming
			// the build is from a buildpack.
			originalCommand = []string{"/lifecycle/entrypoint.bash"}
		}
		originalStartCommand := append(originalCommand, originalArgs...)

		combinedStartCommand := append(fuseCommands, strings.Join(originalStartCommand, " "))

		userContainer.Command = []string{"/bin/sh"}
		userContainer.Args = []string{"-c", strings.Join(combinedStartCommand, " ")}

		userContainer.Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"timeout", "-k", "10s", "10s", "/bin/sh", "-c", strings.Join(unmountCommands, " ")},
				},
			},
		}
	}

	// Populate default pod spec
	spec.RestartPolicy = corev1.RestartPolicyAlways
	spec.TerminationGracePeriodSeconds = space.Status.RuntimeConfig.TerminationGracePeriodSeconds
	spec.DNSPolicy = corev1.DNSClusterFirst

	spec.SecurityContext = &corev1.PodSecurityContext{}
	spec.SchedulerName = corev1.DefaultSchedulerName

	return spec, nil
}

func buildVolumes(volumeStatus []v1alpha1.AppVolumeStatus) ([]corev1.Volume, []corev1.VolumeMount, []string, []string, error) {
	if len(volumeStatus) == 0 {
		return nil, nil, nil, nil, nil
	}

	// Make sure the output is deterministic, volume services shouldn't
	// have dependencies on each other so it should be fine to rearrange.
	sort.Slice(volumeStatus, func(i, j int) bool {
		return volumeStatus[i].MountPath < volumeStatus[j].MountPath
	})

	var volumes []corev1.Volume
	var userVolumeMounts []corev1.VolumeMount
	fuseCommands := []string{}
	unmountCommands := []string{}

	for _, volumeStatus := range volumeStatus {
		fuseMountPath := volumeStatus.MountPath

		nfsVolume := corev1.Volume{
			Name: volumeStatus.VolumeClaimName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: volumeStatus.VolumeClaimName,
				},
			},
		}

		nfsMount := corev1.VolumeMount{
			Name:      volumeStatus.VolumeClaimName,
			MountPath: path.Join(kfMounts, volumeStatus.MountPath),
			ReadOnly:  volumeStatus.ReadOnly,
		}

		volumes = append(volumes, nfsVolume)
		userVolumeMounts = append(userVolumeMounts, nfsMount)

		// Set uid, gid to vcap user if not specified.
		var uid, gid int64
		if volumeStatus.UID == "" {
			uid = buildpackV2VcapUserId
		} else {
			id, err := volumeStatus.UIDInt64()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			uid = id
		}
		if volumeStatus.GID == "" {
			gid = buildpackV2VcapUserId
		} else {
			id, err := volumeStatus.GIDInt64()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			gid = id
		}
		fuseCommands = append(fuseCommands, fmt.Sprintf(mapfsCmdFormat, uid, gid, fuseMountPath, nfsMount.MountPath))
		unmountCommands = append(unmountCommands, fmt.Sprintf(unmountCmdFormat, fuseMountPath))
	}

	unmountCommands = append(unmountCommands, "wait")

	return volumes, userVolumeMounts, fuseCommands, unmountCommands, nil
}

// buildContainerPorts builds a singleton list of ContainerPorts used to connect
// external processes to the app running in the container.
func buildContainerPorts(userPort int32) []corev1.ContainerPort {
	return []corev1.ContainerPort{{
		Name:          UserPortName,
		ContainerPort: userPort,
		Protocol:      corev1.ProtocolTCP,
	}}
}

// rewriteUserProbe adds the detected port to the probe so health checks work.
func rewriteUserProbe(p *corev1.Probe, userPort int32) {
	switch {
	case p == nil:
		return
	case p.HTTPGet != nil:
		p.HTTPGet.Port = intstr.FromInt(int(userPort))
	case p.TCPSocket != nil:
		p.TCPSocket.Port = intstr.FromInt(int(userPort))
	}
}
