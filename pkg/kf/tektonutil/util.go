package tektonutil

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func StringParam(name, description string) tektonv1beta1.ParamSpec {
	return tektonv1beta1.ParamSpec{
		Name:        name,
		Description: description,
		Type:        tektonv1beta1.ParamTypeString,
	}
}
