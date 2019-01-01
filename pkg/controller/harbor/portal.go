package harbor

import (
	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newServiceForPortal(inst *appv1alpha1.Harbor) *v1.Service {
	ls := labelsForHarbor(inst.Name, "portal")
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-portal",
			Namespace: inst.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
			Selector: ls,
		},
	}
	return svc
}

func newPortalForCr(inst *appv1alpha1.Harbor) *appsv1.Deployment {
	ls := labelsForHarbor(inst.Name, "portal")
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-portal",
			Namespace: inst.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "goharbor/harbor-portal:" + inst.Spec.Version,
							Name:  inst.Name + "-portal",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							LivenessProbe: &v1.Probe{
								InitialDelaySeconds: 1,
								PeriodSeconds:       10,
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(80),
									},
								},
							},
							ReadinessProbe: &v1.Probe{
								InitialDelaySeconds: 1,
								PeriodSeconds:       10,
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return dep
}
