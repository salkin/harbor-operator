package harbor

import (
	"bytes"
	"text/template"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newServiceForAdminserver(inst *appv1alpha1.Harbor) *v1.Service {
	ls := labelsForHarbor(inst.Name, "adminserver")
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-adminserver",
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

func newCmForAdminserver(inst *appv1alpha1.Harbor, config *HarborInternal, dir string) *corev1.ConfigMap {

	tmpl := template.Must(template.ParseFiles(dir + "/templates/adminserver.app.conf"))
	buf := bytes.NewBuffer([]byte{})
	err := tmpl.Execute(buf, config)
	if err != nil {
		log.Error(err, "Failed to template adminserver conf")
		return nil
	}
	appData := make(map[string]interface{})
	log.Info("Bytes", "Data", string(buf.Bytes()))
	err = yaml.Unmarshal(buf.Bytes(), &appData)
	if err != nil {
		log.Error(err, "Failed to unmarshal", "appData", appData)
		return nil
	}
	appTemplated := make(map[string]string)
	for k, v := range appData {
		appTemplated[k] = v.(string)
	}
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-adminserver",
			Namespace: inst.Namespace,
		},
		Data: appTemplated,
	}
	return cm
}

func newSecretForAdminserver(inst *appv1alpha1.Harbor, d *HarborInternal) *v1.Secret {
	ls := labelsForHarbor(inst.Name, "adminserver")
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-adminserver",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			"secretKey":             []byte(d.SecretKey),
			"HARBOR_ADMIN_PASSWORD": []byte(d.HarborData.AdminPassword),
			"POSTGRESQL_PASSWORD":   []byte(d.HarborSecrets.DBPassword),
		},
	}
	return sec

}

func newAdminserverForCr(inst *appv1alpha1.Harbor) *appsv1.Deployment {
	ls := labelsForHarbor(inst.Name, "adminserver")
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-adminserver",
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
							Image: "goharbor/harbor-adminserver:" + inst.Spec.Version,
							Name:  inst.Name + "-adminserver",
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "adminserver-key",
									MountPath: "/etc/adminserver/key",
									SubPath:   "key",
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "PORT",
									Value: "8080",
								},
								{
									Name:  "JSON_CFG_STORE_PATH",
									Value: "/etc/adminserver/config/config.json",
								},
								{
									Name:  "KEY_PATH",
									Value: "/etc/adminserver/key",
								},
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
									Name: "JOBSERVICE_SECRET",
									ValueFrom: &v1.EnvVarSource{
										SecretKeyRef: &v1.SecretKeySelector{
											LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-jobservice"},
											Key:                  "secret",
										},
									},
								},
							},
							EnvFrom: []v1.EnvFromSource{
								{
									ConfigMapRef: &v1.ConfigMapEnvSource{
										LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-adminserver"},
									},
								},
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-adminserver"},
									},
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "adminserver-key",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: inst.Name + "-adminserver",
									Items: []v1.KeyToPath{
										{
											Key:  "secretKey",
											Path: "key",
										},
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
