package harbor

import (
	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	postgre "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newStatefulSetForDb(inst *appv1alpha1.Harbor) *appsv1.StatefulSet {
	replicas := int32(1)
	ls := labelsForHarbor(inst.Name, "database")
	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-database",
			Namespace: inst.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: inst.Name + "-database",
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:  "remove-lost-found",
							Image: "goharbor/harbor-db:dev",
							Command: []string{
								"rm", "-Rf", "/var/lib/postgresql/data/lost+found",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "database-data",
									MountPath: "/var/lib/postgresql/data",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "database",
							Image: "goharbor/harbor-db:v1.6.0",
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{Name: inst.Name + "-database"},
									},
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"/docker-healthcheck.sh",
										},
									},
								},
							},
						},
					},
					Volumes: []v1.Volume{},
				},
			},
		},
	}

	v := v1.Volume{
		Name: "database-data",
	}
	if inst.Spec.StorageClass != "" {

		volumeClaim := v1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PersistentVolumeClaim",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: inst.Name + "-database-data",
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteOnce,
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						"storage": resource.MustParse("1Gi"),
					},
				},
			},
		}
		ss.Spec.VolumeClaimTemplates = []v1.PersistentVolumeClaim{volumeClaim}
		v.VolumeSource = v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: inst.Name + "-database-data",
			},
		}
	} else {
		v.VolumeSource = v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		}
	}
	ss.Spec.Template.Spec.Volumes = []v1.Volume{v}

	return ss
}

func newServiceForDb(inst *appv1alpha1.Harbor) *v1.Service {
	ls := labelsForHarbor(inst.Name, "database")
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-database",
			Namespace: inst.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 5432,
				},
			},
			Selector: ls,
		},
	}

	return svc
}

func newSecretForDb(inst *appv1alpha1.Harbor, d *HarborInternal) *v1.Secret {

	data := map[string][]byte{
		"POSTGRES_PASSWORD": []byte(rand.String(16)),
	}
	d.DBPassword = string(data["POSTGRES_PASSWORD"])
	ls := labelsForHarbor(inst.Name, "database")
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-database",
			Namespace: inst.Namespace,
			Labels:    ls,
		},
		Data: data,
	}
	return sec
}
func newCmForDb(inst *appv1alpha1.Harbor) *v1.ConfigMap {
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

func newPostgreCrd(inst *appv1alpha1.Harbor) *postgre.Postgresql {
	cr := &postgre.Postgresql{
		TypeMeta: metav1.TypeMeta{
			Kind:       "postgresql",
			APIVersion: "acid.zalan.do/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name + "-postgre",
			Namespace: inst.Namespace,
		},
		Spec: postgre.PostgresSpec{
			TeamID: inst.Name,
			Volume: postgre.Volume{
				Size: "1Gi",
			},
			NumberOfInstances: 1,
			Users: map[string]postgre.UserFlags{
				"harbor": {"superuser"},
			},
			Databases: map[string]string{
				"harbor": "harbor",
			},

			PostgresqlParam: postgre.PostgresqlParam{},
		},
	}
	return cr
}
