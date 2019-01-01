package harbor

import (
	"bytes"
	"text/template"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	coreCm = "-core"
)

func newPVCForRegistry(inst *appv1alpha1.Harbor) *v1.PersistentVolumeClaim {
	ls := labelsForHarbor(inst.Name, "registry")
	volumeClaim := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-registry",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": resource.MustParse("10Gi"),
				},
			},
		},
	}

	if inst.Spec.Registry.Storage.StorageClass != "" {
		volumeClaim.Spec.StorageClassName = &inst.Spec.Registry.Storage.StorageClass
	}

	if inst.Spec.Registry.Storage.Size != "" {
		volumeClaim.Spec.Resources.Requests["storage"] = resource.MustParse(inst.Spec.Registry.Storage.Size)

	}

	return volumeClaim
}

func newServiceForRegistry(inst *appv1alpha1.Harbor) *v1.Service {
	ls := labelsForHarbor(inst.Name, "registry")
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-registry",
			Namespace: inst.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 5000,
					Name: "registry",
				},
				{
					Port: 8080,
					Name: "controller",
				},
			},
			Selector: ls,
		},
	}
	return svc
}

func newCmForRegistry(inst *appv1alpha1.Harbor, config *HarborInternal, dir string) *v1.ConfigMap {
	t := template.Must(template.ParseFiles(dir + "/templates/registry-cm.yml"))
	b := bytes.NewBuffer([]byte{})
	err := t.Execute(b, config)
	if err != nil {
		log.Error(err, "Failed to template registry cm")
		panic("Failed to template" + err.Error())
	}
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-registry",
			Namespace: inst.Namespace,
		},
		Data: map[string]string{
			"config.yml": string(b.Bytes()),
		},
	}
	return cm

}

func newRegistryForCr(inst *appv1alpha1.Harbor) *appsv1.Deployment {
	ls := labelsForHarbor(inst.Name, "registry")
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-registry",
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
							Image: "goharbor/registry-photon:dev",
							Name:  inst.Name + "-core",
							Args: []string{
								"serve", "/etc/registry/config.yml",
							},
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-registry"},
									},
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "storage",
									MountPath: "/storage",
								},
								{
									Name:      "registry-config",
									MountPath: "/etc/registry/config.yml",
									SubPath:   "config.yml",
								},
								{
									Name:      "registry-root-certificate",
									MountPath: "/etc/registry/root.crt",
									SubPath:   "tokenServiceRootCertBundle",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "registry-root-certificate",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: inst.Name + "-core",
								},
							},
						},
						{
							Name: "registry-config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-registry"},
								},
							},
						},
					},
				},
			},
		},
	}

	storage := v1.Volume{}
	if inst.Spec.Registry.Storage.StorageClass != "" {
		storage = v1.Volume{
			Name: "storage",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: inst.Name + "-registry",
				},
			},
		}
	} else {
		storage = v1.Volume{
			Name: "storage",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}
	}
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, storage)
	return dep
}

func newSecretForRegistry(inst *appv1alpha1.Harbor) *v1.Secret {

	ls := labelsForHarbor(inst.Name, "core")
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-registry",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{
			"REGISTRY_HTTP_SECRET": []byte(rand.String(16)),
		},
	}
	return sec
}

func labelsForHarbor(appLabel string, component string) map[string]string {
	return map[string]string{"app": appLabel, "component": component}
}
