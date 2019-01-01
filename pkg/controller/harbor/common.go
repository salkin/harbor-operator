package harbor

import (
	"encoding/json"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// HarborInternal Holds data internal to an Harbor instance
type HarborInternal struct {
	HarborData    `json:"harbor"`
	Clair         `json:"clair,omitempty"`
	Notary        `json:"notary"`
	HarborSecrets `json:"secrets"`
}

type Notary struct {
	Enabled bool
}

type HarborSecrets struct {
	DBPassword string
	DBUser     string
}

type HarborData struct {
	// Name of the CR instance
	Name string
	Storage
	// AdminPassword used in portal
	AdminPassword string `json:"adminpassword"`
	ExtEndpoint   string `json:"externalUrl"`
	CoreURL       string `json:"coreUrl"`
	JobserviceURL string `json:"jobUrl"`
	LogLevel      string `json:"logLevel"`
	DBURL         string `json:"dbURL"`
	SecretKey     string
	Database
	Chart
	Notary
}

type Chart struct {
	Enabled     string
	StorageType string
}

type Storage struct {
	Type string
}

type Database struct {
	Sslmode       string
	CoreDatabase  string
	ClairDatabase string
}

type Clair struct {
	Enabled string
}

type Postgresq struct {
}

func (r *ReconcileHarbor) setReference(cr *appv1alpha1.Harbor, obj metav1.Object) error {
	// Set Harbor instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, obj, r.scheme); err != nil {
		return err
	}
	return nil
}

func createNewHarborData(cr *appv1alpha1.Harbor) *HarborInternal {
	logLevel := "info"
	if cr.Spec.Config.LogLevel != "" {
		logLevel = cr.Spec.Config.LogLevel
	}
	data := &HarborInternal{
		HarborData: HarborData{
			Name: cr.Name,
			Storage: Storage{
				Type: "filesystem",
			},
			LogLevel:      logLevel,
			SecretKey:     rand.String(16),
			AdminPassword: rand.String(16),
			ExtEndpoint:   cr.Spec.Config.ExtURL,
			CoreURL:       cr.Name + "-core",
		},
	}
	return data
}

func newCmForHarborInt(cr *appv1alpha1.Harbor) *v1.ConfigMap {
	intConfig := createNewHarborData(cr)
	st, _ := json.Marshal(intConfig)
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-harbor-int",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"data.json": string(st),
		},
	}
	return cm
}
