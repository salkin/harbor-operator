package harbor

import (
	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type createFunc func(*appv1alpha1.Harbor, runtime.Object) error
type deployFunc func(*appv1alpha1.Harbor) error

type errCreator struct {
	err error
	c   createFunc
}

func (e *errCreator) create(cr *appv1alpha1.Harbor, obj runtime.Object) {
	if e.err != nil {
		return
	}
	e.err = e.c(cr, obj)
}
