package harbor

import (
	"strings"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	"k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newIngressForCR(inst *appv1alpha1.Harbor) *extv1beta1.Ingress {
	host := strings.Replace(inst.Spec.Config.ExtURL, "http://", "", -1)
	ing := &extv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-ing",
			Namespace: inst.Namespace,
		},
		Spec: extv1beta1.IngressSpec{
			Rules: []extv1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: extv1beta1.IngressRuleValue{
						HTTP: &extv1beta1.HTTPIngressRuleValue{
							Paths: []extv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: inst.Name + "-portal",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/c/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: inst.Name + "-core",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/chartrepo/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: inst.Name + "-core",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/v2/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: inst.Name + "-core",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/service/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: inst.Name + "-core",
										ServicePort: intstr.FromInt(80),
									},
								},
								{
									Path: "/api/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: inst.Name + "-core",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ing
}

func newSecretForIngress(inst *appv1alpha1.Harbor) *v1.Secret {

	names := []string{"notary." + inst.Spec.Config.ExtURL, "registry." + inst.Spec.Config.ExtURL}
	rootCa, rootKey := createNewRoot(names)
	servPem, servKey := createNewCertificate(rootKey)
	servKeyPem := pemFromKey(servKey)
	ls := labelsForHarbor(inst.Name, "core")
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-ingress",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			"tls.crt": servPem,
			"tls.key": servKeyPem,
			"ca.crt":  rootCa,
		},
	}
	return sec

}
