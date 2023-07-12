package tektonutil

import (
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
)

func StringParam(name, description string) tektonv1.ParamSpec {
	return tektonv1.ParamSpec{
		Name:        name,
		Description: description,
		Type:        tektonv1.ParamTypeString,
	}
}

func DefaultStringParam(name, description, defaultValue string) tektonv1.ParamSpec {
	out := StringParam(name, description)
	out.Default = &tektonv1.ParamValue{
		Type:      tektonv1.ParamTypeString,
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
