package harbor

import (
	"testing"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestTemplateCm(t *testing.T) {
	cr := &appv1alpha1.Harbor{}
	config := HarborInternal{
		HarborData: HarborData{
			LogLevel: "info",
		},
	}
	cm := newCmForRegistry(cr, &config, "../../../")
	assert.NotNil(t, cm)
	data := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(cm.Data["config.yml"]), &data)
	assert.Nil(t, err)
	d := data["notifications"]
	t.Logf("Data %v", data)
	assert.NotNil(t, d)

}
