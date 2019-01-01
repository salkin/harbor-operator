package harbor

import (
	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newServiceForCR(inst *appv1alpha1.Harbor) *v1.Service {
	ls := labelsForHarbor(inst.Name, "core")
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-core",
			Namespace: inst.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: ls,
		},
	}
	return svc
}

// newCoreForCR creates the deployment for Core Deployment/Pod
func newCoreForCR(inst *appv1alpha1.Harbor) *appsv1.Deployment {
	ls := labelsForHarbor(inst.Name, "core")
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-core",
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
							Image: "goharbor/harbor-core:" + inst.Spec.Version,
							Name:  inst.Name + "-core",
							Env: []v1.EnvVar{
								{
									Name: "CORE_SECRET",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-core"},
											Key:                  "secret",
										},
									},
								},
								{
									Name:  "ADMINSERVER_URL",
									Value: inst.Name + "-adminserver",
								},
								{
									Name:  "SYNC_REGISTRY",
									Value: "false",
								},
								{
									Name:  "CONFIG_PATH",
									Value: "/etc/core/app.conf",
								},
								{
									Name:  "CHART_CACHE_DRIVER",
									Value: "redis",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/core/app.conf",
									SubPath:   "app.conf",
								},
								{
									Name:      "secret-key",
									MountPath: "/etc/core/key",
									SubPath:   "key",
								},
								{
									Name:      "ca-download",
									MountPath: "/etc/core/ca/ca.crt",
									SubPath:   "ca.crt",
								},
								{
									Name:      "token-service-private-key",
									MountPath: "/etc/core/private_key.pem",
									SubPath:   "tokenServicePrivateKey",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-core"},
								},
							},
						},
						{
							Name: "secret-key",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: inst.Name + "-core",
									Items: []v1.KeyToPath{
										{
											Key:  "secretKey",
											Path: "key",
										},
									},
								},
							},
						},
						{
							Name: "token-service-private-key",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: inst.Name + "-core",
								},
							},
						},
						{
							Name: "ca-download",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: inst.Name + "-ingress",
								},
							},
						},
						{
							Name: "psc",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	return dep
}

func newSecretForCore(inst *appv1alpha1.Harbor, d *HarborInternal) *v1.Secret {

	rootCa, rootKey := createHarbRoot()
	ls := labelsForHarbor(inst.Name, "core")
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-core",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			"secretKey":                  []byte(d.SecretKey),
			"secret":                     []byte(rand.String(16)),
			"tokenServiceRootCertBundle": rootCa,
			"tokenServicePrivateKey":     rootKey,
		},
	}
	return sec
}

func newCoreCmForCR(inst *appv1alpha1.Harbor) *v1.ConfigMap {
	//TODO: make Appconfig dynamically templated
	appConfig := `appname = Harbor
	runmode = prod
	enablegzip = true

	[prod]
	httpport = 8080`

	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-core",
			Namespace: inst.Namespace,
		},
		Data: map[string]string{
			"app.conf": appConfig,
		},
	}
	return cm
}
