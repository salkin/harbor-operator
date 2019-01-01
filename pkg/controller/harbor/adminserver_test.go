package harbor

import (
	"testing"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestAdminCm(t *testing.T) {

	initLog()
	cr := &appv1alpha1.Harbor{}
	config := HarborInternal{}
	cm := newCmForAdminserver(cr, &config, "../../..")

	assert.NotNil(t, cm)

}
