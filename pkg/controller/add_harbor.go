package controller

import (
	"github.com/salkin/harbor-operator/pkg/controller/harbor"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, harbor.Add)
}
