package harbor

import (
	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newSecretForJobservice(inst *appv1alpha1.Harbor) *v1.Secret {

	ls := labelsForHarbor(inst.Name, "jobservice")
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-jobservice",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			"secret": []byte(rand.String(16)),
		},
	}
	return sec
}

func newCmForJobservice(inst *appv1alpha1.Harbor) *v1.ConfigMap {
	config := ``
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-jobservice",
			Namespace: inst.Namespace,
		},
		Data: map[string]string{
			"config.yaml": config,
		},
	}
	return cm
}
