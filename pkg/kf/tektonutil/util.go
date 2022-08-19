package tektonutil

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func StringParam(name, description string) tektonv1beta1.ParamSpec {
	return tektonv1beta1.ParamSpec{
		Name:        name,
		Description: description,
		Type:        tektonv1beta1.ParamTypeString,
	}
}

func DefaultStringParam(name, description, defaultValue string) tektonv1beta1.ParamSpec {
	out := StringParam(name, description)
	out.Default = &tektonv1beta1.ArrayOrString{
		Type:      tektonv1beta1.ParamTypeString,
		StringVal: defaultValue,
	}

	return out
}

func EmptyVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}
