// Copyright 2019 Google LLC

package resources

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/kf/v2/pkg/apis/kf/v1alpha1"
	"github.com/google/kf/v2/pkg/kf/testutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/ptr"
)

func ExampleDeploymentName() {
	app := &v1alpha1.App{}
	app.Name = "my-app"

	fmt.Println("Deployment name:", DeploymentName(app))

	// Output: Deployment name: my-app
}

func TestMakeDeployment(t *testing.T) {
	tests := map[string]struct {
		app     *v1alpha1.App
		space   *v1alpha1.Space
		want    *appsv1.Deployment
		wantErr error
	}{
		"missing image": {
			app:     &v1alpha1.App{},
			wantErr: errors.New("waiting for build image in latestReadyBuild"),
		},
		"missing replicas": {
			app: &v1alpha1.App{
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "gcr.io/my-app",
					},
				},
			},
			wantErr: errors.New("Exact scale required for deployment based setup"),
		},
		"stopped": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Stopped:  true,
						Replicas: ptr.Int32(30),
					},
				},
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "gcr.io/my-app",
					},
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						ProgressDeadlineSeconds: ptr.Int32(600),
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "app-scaler",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "my-app",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "kf.dev/v1alpha1",
							Kind:               "App",
							Name:               "my-app",
							Controller:         ptr.Bool(true),
							BlockOwnerDeletion: ptr.Bool(true),
						},
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas:             ptr.Int32(0),
					RevisionHistoryLimit: ptr.Int32(10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/component":  "app-server",
							"app.kubernetes.io/managed-by": "kf",
							"app.kubernetes.io/name":       "my-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/component":  "app-server",
								"app.kubernetes.io/managed-by": "kf",
								"app.kubernetes.io/name":       "my-app",
								v1alpha1.NetworkPolicyLabel:    v1alpha1.NetworkPolicyApp,
							},
							Annotations: map[string]string{
								"sidecar.istio.io/inject":                          "true",
								"traffic.sidecar.istio.io/includeOutboundIPRanges": "*",
								"kubectl.kubernetes.io/default-container":          "user-container",
							},
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &appsv1.RollingUpdateDeployment{
							MaxUnavailable: &defaultMaxUnavailable,
							MaxSurge:       &defaultMaxSurge,
						},
					},
					ProgressDeadlineSeconds: ptr.Int32(600),
				},
			},
		},
		"app-labels": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
					Labels: map[string]string{
						"user-label1": "label1-value",
						"user-label2": "label2-value",
					},
					Annotations: map[string]string{
						"user-annotation1": "annotation1-value",
						"user-annotation2": "annotation2-value",
					},
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Stopped:  true,
						Replicas: ptr.Int32(30),
					},
				},
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "gcr.io/my-app",
					},
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						ProgressDeadlineSeconds: ptr.Int32(600),
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "app-scaler",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "my-app",
						"user-label1":                  "label1-value",
						"user-label2":                  "label2-value",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "kf.dev/v1alpha1",
							Kind:               "App",
							Name:               "my-app",
							Controller:         ptr.Bool(true),
							BlockOwnerDeletion: ptr.Bool(true),
						},
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas:             ptr.Int32(0),
					RevisionHistoryLimit: ptr.Int32(10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/component":  "app-server",
							"app.kubernetes.io/managed-by": "kf",
							"app.kubernetes.io/name":       "my-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"user-label1":                  "label1-value",
								"user-label2":                  "label2-value",
								"app.kubernetes.io/component":  "app-server",
								"app.kubernetes.io/managed-by": "kf",
								"app.kubernetes.io/name":       "my-app",
								v1alpha1.NetworkPolicyLabel:    v1alpha1.NetworkPolicyApp,
							},
							Annotations: map[string]string{
								"sidecar.istio.io/inject":                          "true",
								"traffic.sidecar.istio.io/includeOutboundIPRanges": "*",
								"kubectl.kubernetes.io/default-container":          "user-container",
								"user-annotation1":                                 "annotation1-value",
								"user-annotation2":                                 "annotation2-value",
							},
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &appsv1.RollingUpdateDeployment{
							MaxUnavailable: &defaultMaxUnavailable,
							MaxSurge:       &defaultMaxSurge,
						},
					},
					ProgressDeadlineSeconds: ptr.Int32(600),
				},
			},
		},
		"app-labels-lowest-priority": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "user-defined",
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: v1alpha1.AppSpec{
					Instances: v1alpha1.AppSpecInstances{
						Stopped:  true,
						Replicas: ptr.Int32(30),
					},
				},
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "gcr.io/my-app",
					},
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						ProgressDeadlineSeconds: ptr.Int32(600),
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
					Labels: map[string]string{
						"app.kubernetes.io/component":  "app-scaler",
						"app.kubernetes.io/managed-by": "kf",
						"app.kubernetes.io/name":       "my-app",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "kf.dev/v1alpha1",
							Kind:               "App",
							Name:               "my-app",
							Controller:         ptr.Bool(true),
							BlockOwnerDeletion: ptr.Bool(true),
						},
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas:             ptr.Int32(0),
					RevisionHistoryLimit: ptr.Int32(10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/component":  "app-server",
							"app.kubernetes.io/managed-by": "kf",
							"app.kubernetes.io/name":       "my-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/component":  "app-server",
								"app.kubernetes.io/managed-by": "kf",
								"app.kubernetes.io/name":       "my-app",
								v1alpha1.NetworkPolicyLabel:    v1alpha1.NetworkPolicyApp,
							},
							Annotations: map[string]string{
								"sidecar.istio.io/inject":                          "true",
								"traffic.sidecar.istio.io/includeOutboundIPRanges": "*",
								"kubectl.kubernetes.io/default-container":          "user-container",
							},
						},
					},
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &appsv1.RollingUpdateDeployment{
							MaxUnavailable: &defaultMaxUnavailable,
							MaxSurge:       &defaultMaxSurge,
						},
					},
					ProgressDeadlineSeconds: ptr.Int32(600),
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// automatically fill in desired spec
			if tc.want != nil {
				podSpec, _ := makePodSpec(tc.app, tc.space)
				tc.want.Spec.Template.Spec = *podSpec
			}
			got, err := MakeDeployment(tc.app, tc.space)
			testutil.AssertEqual(t, "Deployment", tc.want, got)
			testutil.AssertEqual(t, "Error", tc.wantErr, err)
		})
	}
}

func Test_makePodSpec(t *testing.T) {
	tests := map[string]struct {
		app   *v1alpha1.App
		space *v1alpha1.Space

		want func(app *v1alpha1.App) corev1.PodSpec
	}{
		"default": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						TerminationGracePeriodSeconds: ptr.Int64(30),
					},
				},
			},
			want: func(app *v1alpha1.App) corev1.PodSpec {
				var wantEnv []corev1.EnvVar

				wantEnv = append(wantEnv, BuildRuntimeEnvVars(CFRunning, app)...)
				wantEnv = append(wantEnv, corev1.EnvVar{Name: "KF_UPDATE_REQUESTS_", Value: "0"})

				return corev1.PodSpec{
					EnableServiceLinks: ptr.Bool(false),
					Containers: []corev1.Container{
						{
							Name:                     "user-container",
							Ports:                    buildContainerPorts(DefaultUserPort),
							Env:                      wantEnv,
							Stdin:                    false,
							TTY:                      false,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						},
					},
					NodeSelector:                  map[string]string{},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: ptr.Int64(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 corev1.DefaultSchedulerName,
				}
			},
		},
		"populated": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Template: v1alpha1.AppSpecTemplate{
						UpdateRequests: 10,
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "name-override",
									Env: []corev1.EnvVar{
										{Name: "app-key", Value: "bar"},
									},
									Ports: []corev1.ContainerPort{
										{Name: "http-user", ContainerPort: 9999},
										{Name: "http-admin", ContainerPort: 5000},
									},
									LivenessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/v1/healthz",
											},
										},
									},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											TCPSocket: &corev1.TCPSocketAction{},
										},
									},
									StartupProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/v1/livez",
											},
										},
									},
								},
							},
						},
					},
					Build: v1alpha1.AppSpecBuild{
						Spec: &v1alpha1.BuildSpec{
							NodeSelector: map[string]string{
								"disktype": "ssd10",
							},
						},
					},
				},
				Status: v1alpha1.AppStatus{
					BuildStatusFields: v1alpha1.BuildStatusFields{
						Image: "gcr.io/my-app",
					},
					ServiceAccountName: "sa-my-app",
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						Env: []corev1.EnvVar{
							{Name: "space-key", Value: "bar"},
						},
						TerminationGracePeriodSeconds: ptr.Int64(30),
					},
				},
				Spec: v1alpha1.SpaceSpec{
					RuntimeConfig: v1alpha1.SpaceSpecRuntimeConfig{
						NodeSelector: map[string]string{
							"disktype": "ssd",
							"cpu":      "amd64",
						},
					},
				},
			},

			want: func(app *v1alpha1.App) corev1.PodSpec {
				wantEnv := []corev1.EnvVar{
					// envs must cascade space -> app -> kf
					{Name: "space-key", Value: "bar"},
					{Name: "app-key", Value: "bar"},
				}

				wantEnv = append(wantEnv, BuildRuntimeEnvVars(CFRunning, app)...)
				wantEnv = append(wantEnv, corev1.EnvVar{Name: "KF_UPDATE_REQUESTS_", Value: "10"})

				return corev1.PodSpec{
					EnableServiceLinks: ptr.Bool(false),
					Containers: []corev1.Container{
						{
							Name:  "user-container", // remains user-container if overridden
							Image: "gcr.io/my-app",  // copied from status
							Ports: []corev1.ContainerPort{ // container ports preserved
								{Name: "http-user", ContainerPort: 9999},
								{Name: "http-admin", ContainerPort: 5000},
							},
							Env:   wantEnv,
							Stdin: false,
							TTY:   false,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/v1/healthz",
										Port: intstr.FromInt(9999),
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromInt(9999),
									},
								},
							},
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/v1/livez",
										Port: intstr.FromInt(9999),
									},
								},
							},
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						},
					},
					NodeSelector: map[string]string{
						"disktype": "ssd10",
						"cpu":      "amd64",
					},
					ServiceAccountName:       ServiceAccountName(app),
					DeprecatedServiceAccount: ServiceAccountName(app),

					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: ptr.Int64(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 corev1.DefaultSchedulerName,
				}
			},
		},

		"nfs volume default app": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Template: v1alpha1.AppSpecTemplate{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "gcr.io/my-app",
								},
							},
						},
					},
				},
				Status: v1alpha1.AppStatus{
					Volumes: []v1alpha1.AppVolumeStatus{
						{
							MountPath:       "/nfs/volume1",
							VolumeName:      "nfsvolume1",
							ReadOnly:        false,
							VolumeClaimName: "nfs-volume1-pvc",
						},
					},
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						TerminationGracePeriodSeconds: ptr.Int64(30),
					},
				},
			},
			want: func(app *v1alpha1.App) corev1.PodSpec {
				var wantEnv []corev1.EnvVar

				wantEnv = append(wantEnv, BuildRuntimeEnvVars(CFRunning, app)...)
				wantEnv = append(wantEnv, corev1.EnvVar{Name: "KF_UPDATE_REQUESTS_", Value: "0"})

				return corev1.PodSpec{
					EnableServiceLinks: ptr.Bool(false),
					Containers: []corev1.Container{
						{
							Name:                     "user-container",
							Ports:                    buildContainerPorts(DefaultUserPort),
							Env:                      wantEnv,
							Stdin:                    false,
							TTY:                      false,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							Command:                  []string{"/bin/sh"},
							Args: []string{
								"-c",
								"mapfs -uid 2000 -gid 2000 /nfs/volume1 /.kfmounts/nfs/volume1 & exec /lifecycle/entrypoint.bash",
							},
							VolumeMounts: []corev1.VolumeMount{{Name: "nfs-volume1-pvc", MountPath: "/.kfmounts/nfs/volume1"}},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"timeout", "-k", "10s", "10s", "/bin/sh", "-c", "fusermount -u -z /nfs/volume1 & wait"},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.Bool(true),
							},
						},
					},
					NodeSelector:                  map[string]string{},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: ptr.Int64(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 corev1.DefaultSchedulerName,
					Volumes: []corev1.Volume{
						{
							Name: "nfs-volume1-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nfs-volume1-pvc",
								},
							},
						},
					},
				}
			},
		},

		"nfs volume app with args": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Template: v1alpha1.AppSpecTemplate{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "gcr.io/my-app",
									Args:  []string{"-jar", "my-library.jar", "-timeout=10", "-needsShellEscape=\"$'"},
								},
							},
						},
					},
				},
				Status: v1alpha1.AppStatus{
					Volumes: []v1alpha1.AppVolumeStatus{
						{
							MountPath:       "/nfs/volume1",
							VolumeName:      "nfsvolume1",
							ReadOnly:        false,
							VolumeClaimName: "nfs-volume1-pvc",
						},
					},
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						TerminationGracePeriodSeconds: ptr.Int64(30),
					},
				},
			},
			want: func(app *v1alpha1.App) corev1.PodSpec {
				var wantEnv []corev1.EnvVar

				wantEnv = append(wantEnv, BuildRuntimeEnvVars(CFRunning, app)...)
				wantEnv = append(wantEnv, corev1.EnvVar{Name: "KF_UPDATE_REQUESTS_", Value: "0"})

				return corev1.PodSpec{
					EnableServiceLinks: ptr.Bool(false),
					Containers: []corev1.Container{
						{
							Name:                     "user-container",
							Ports:                    buildContainerPorts(DefaultUserPort),
							Env:                      wantEnv,
							Stdin:                    false,
							TTY:                      false,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							Command:                  []string{"/bin/sh"},
							Args: []string{
								"-c",
								`mapfs -uid 2000 -gid 2000 /nfs/volume1 /.kfmounts/nfs/volume1 & exec /lifecycle/entrypoint.bash -jar my-library.jar -timeout=10 '-needsShellEscape="$'"'"''`,
							},
							VolumeMounts: []corev1.VolumeMount{{Name: "nfs-volume1-pvc", MountPath: "/.kfmounts/nfs/volume1"}},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"timeout", "-k", "10s", "10s", "/bin/sh", "-c", "fusermount -u -z /nfs/volume1 & wait"},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.Bool(true),
							},
						},
					},
					NodeSelector:                  map[string]string{},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: ptr.Int64(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 corev1.DefaultSchedulerName,
					Volumes: []corev1.Volume{
						{
							Name: "nfs-volume1-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nfs-volume1-pvc",
								},
							},
						},
					},
				}
			},
		},

		"nfs volume app with args and command": {
			app: &v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: v1alpha1.AppSpec{
					Template: v1alpha1.AppSpecTemplate{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image:   "gcr.io/my-app",
									Args:    []string{"-jar", "my-library.jar", "-timeout=10"},
									Command: []string{"java"},
								},
							},
						},
					},
				},
				Status: v1alpha1.AppStatus{
					Volumes: []v1alpha1.AppVolumeStatus{
						{
							MountPath:       "/nfs/volume1",
							VolumeName:      "nfsvolume1",
							ReadOnly:        false,
							VolumeClaimName: "nfs-volume1-pvc",
						},
					},
				},
			},
			space: &v1alpha1.Space{
				Status: v1alpha1.SpaceStatus{
					RuntimeConfig: v1alpha1.SpaceStatusRuntimeConfig{
						TerminationGracePeriodSeconds: ptr.Int64(30),
					},
				},
			},
			want: func(app *v1alpha1.App) corev1.PodSpec {
				var wantEnv []corev1.EnvVar

				wantEnv = append(wantEnv, BuildRuntimeEnvVars(CFRunning, app)...)
				wantEnv = append(wantEnv, corev1.EnvVar{Name: "KF_UPDATE_REQUESTS_", Value: "0"})

				return corev1.PodSpec{
					EnableServiceLinks: ptr.Bool(false),
					Containers: []corev1.Container{
						{
							Name:                     "user-container",
							Ports:                    buildContainerPorts(DefaultUserPort),
							Env:                      wantEnv,
							Stdin:                    false,
							TTY:                      false,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							Command:                  []string{"/bin/sh"},
							Args: []string{
								"-c",
								"mapfs -uid 2000 -gid 2000 /nfs/volume1 /.kfmounts/nfs/volume1 & exec java -jar my-library.jar -timeout=10",
							},
							VolumeMounts: []corev1.VolumeMount{{Name: "nfs-volume1-pvc", MountPath: "/.kfmounts/nfs/volume1"}},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"timeout", "-k", "10s", "10s", "/bin/sh", "-c", "fusermount -u -z /nfs/volume1 & wait"},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.Bool(true),
							},
						},
					},
					NodeSelector:                  map[string]string{},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: ptr.Int64(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 corev1.DefaultSchedulerName,
					Volumes: []corev1.Volume{
						{
							Name: "nfs-volume1-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nfs-volume1-pvc",
								},
							},
						},
					},
				}
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {

			got, _ := makePodSpec(tc.app, tc.space)
			testutil.AssertEqual(t, "PodSpec", tc.want(tc.app), *got)
		})
	}
}

func Test_buildVolumes(t *testing.T) {

	var gid v1alpha1.ID = "2002"
	var uid v1alpha1.ID = "2001"
	tests := map[string]struct {
		volumeStatus []v1alpha1.AppVolumeStatus

		wantVolumes         []corev1.Volume
		wantMounts          []corev1.VolumeMount
		wantFuseCommands    []string
		wantUnmountCommands []string
		wantError           error
	}{
		"nil volumes": {
			volumeStatus:        nil,
			wantVolumes:         nil,
			wantMounts:          nil,
			wantFuseCommands:    nil,
			wantUnmountCommands: nil,
		},
		"empty volumes": {
			volumeStatus:        []v1alpha1.AppVolumeStatus{},
			wantVolumes:         nil,
			wantMounts:          nil,
			wantFuseCommands:    nil,
			wantUnmountCommands: nil,
		},
		"No GID or UID": {
			volumeStatus: []v1alpha1.AppVolumeStatus{
				{
					VolumeName:      "volumeName",
					VolumeClaimName: "volumeClaimName",
					MountPath:       "/mount/path",
				},
			},

			wantVolumes: []corev1.Volume{
				{
					Name: "volumeClaimName",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "volumeClaimName",
						},
					},
				},
			},
			wantMounts: []corev1.VolumeMount{
				{Name: "volumeClaimName", MountPath: "/.kfmounts/mount/path"},
			},
			wantFuseCommands:    []string{"mapfs -uid 2000 -gid 2000 /mount/path /.kfmounts/mount/path &"},
			wantUnmountCommands: []string{"fusermount -u -z /mount/path &", "wait"},
			wantError:           nil,
		},
		"Valid GID and UID": {
			volumeStatus: []v1alpha1.AppVolumeStatus{
				{
					VolumeName:      "volumeName",
					VolumeClaimName: "volumeClaimName",
					MountPath:       "/mount/path",
					UidGid: v1alpha1.UidGid{
						UID: uid,
						GID: gid,
					},
				},
			},
			wantVolumes: []corev1.Volume{
				{
					Name: "volumeClaimName",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "volumeClaimName",
						},
					},
				},
			},
			wantMounts: []corev1.VolumeMount{
				{Name: "volumeClaimName", MountPath: "/.kfmounts/mount/path"},
			},
			wantFuseCommands:    []string{"mapfs -uid 2001 -gid 2002 /mount/path /.kfmounts/mount/path &"},
			wantUnmountCommands: []string{"fusermount -u -z /mount/path &", "wait"},
		},
		"Valid GID and UID, multiple volumes": {
			volumeStatus: []v1alpha1.AppVolumeStatus{
				{
					VolumeName:      "volume1",
					VolumeClaimName: "volumeClaim1",
					MountPath:       "/mount/path1",
					UidGid: v1alpha1.UidGid{
						UID: uid,
						GID: gid,
					},
				},
				{
					VolumeName:      "volume2",
					VolumeClaimName: "volumeClaim2",
					MountPath:       "/mount/path2",
					UidGid: v1alpha1.UidGid{
						UID: uid,
						GID: gid,
					},
				},
			},
			wantVolumes: []corev1.Volume{
				{
					Name: "volumeClaim1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "volumeClaim1",
						},
					},
				},
				{
					Name: "volumeClaim2",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "volumeClaim2",
						},
					},
				},
			},
			wantMounts: []corev1.VolumeMount{
				{Name: "volumeClaim1", MountPath: "/.kfmounts/mount/path1"},
				{Name: "volumeClaim2", MountPath: "/.kfmounts/mount/path2"},
			},
			wantFuseCommands: []string{
				"mapfs -uid 2001 -gid 2002 /mount/path1 /.kfmounts/mount/path1 &",
				"mapfs -uid 2001 -gid 2002 /mount/path2 /.kfmounts/mount/path2 &",
			},
			wantUnmountCommands: []string{"fusermount -u -z /mount/path1 &", "fusermount -u -z /mount/path2 &", "wait"},
		},
		"volumes sorted by path": {
			volumeStatus: []v1alpha1.AppVolumeStatus{
				{
					VolumeName:      "volume2",
					VolumeClaimName: "volumeClaim2",
					MountPath:       "/mount/path2",
				},
				{
					VolumeName:      "volume1",
					VolumeClaimName: "volumeClaim1",
					MountPath:       "/mount/path1",
				},
			},
			wantVolumes: []corev1.Volume{
				{
					Name: "volumeClaim1",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "volumeClaim1",
						},
					},
				},
				{
					Name: "volumeClaim2",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "volumeClaim2",
						},
					},
				},
			},
			wantMounts: []corev1.VolumeMount{
				{Name: "volumeClaim1", MountPath: "/.kfmounts/mount/path1"},
				{Name: "volumeClaim2", MountPath: "/.kfmounts/mount/path2"},
			},
			wantFuseCommands: []string{
				"mapfs -uid 2000 -gid 2000 /mount/path1 /.kfmounts/mount/path1 &",
				"mapfs -uid 2000 -gid 2000 /mount/path2 /.kfmounts/mount/path2 &",
			},
			wantUnmountCommands: []string{"fusermount -u -z /mount/path1 &", "fusermount -u -z /mount/path2 &", "wait"},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			volumes, mounts, fuseCommands, unmountCommands, err := buildVolumes(tc.volumeStatus)

			testutil.AssertEqual(t, "error", tc.wantError, err)
			testutil.AssertEqual(t, "volumes", tc.wantVolumes, volumes)
			testutil.AssertEqual(t, "mounts", tc.wantMounts, mounts)
			testutil.AssertEqual(t, "fuseCommands", tc.wantFuseCommands, fuseCommands)
			testutil.AssertEqual(t, "unmountCommands", tc.wantUnmountCommands, unmountCommands)
		})
	}
}

func Test_buildContainerPorts(t *testing.T) {
	tests := map[string]struct {
		userPort int32
		want     []corev1.ContainerPort
	}{
		"custom": {
			userPort: 300,
			want: []corev1.ContainerPort{
				{
					ContainerPort: 300,
					Name:          UserPortName,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
		"default": {
			userPort: DefaultUserPort,
			want: []corev1.ContainerPort{
				{
					ContainerPort: DefaultUserPort,
					Name:          UserPortName,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			actual := buildContainerPorts(tc.userPort)
			testutil.AssertEqual(t, "containerport", tc.want, actual)
		})
	}
}

func Test_rewriteUserProbe(t *testing.T) {
	tests := map[string]struct {
		probe    *corev1.Probe
		userPort int32

		want *corev1.Probe
	}{
		"nil probe": {
			probe:    nil,
			userPort: 2000,
			want:     nil,
		},
		"HTTP probe": {
			probe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/foo",
					},
				},
			},
			userPort: 3000,
			want: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/foo",
						Port: intstr.FromInt(3000),
					},
				},
			},
		},
		"TCPSocket probe": {
			probe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
			userPort: 3000,
			want: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(3000)},
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			rewriteUserProbe(tc.probe, tc.userPort)
			testutil.AssertEqual(t, "probe", tc.want, tc.probe)
		})
	}
}
