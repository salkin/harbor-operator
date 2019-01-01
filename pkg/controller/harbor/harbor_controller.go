package harbor

import (
	"context"
	"encoding/json"
	"fmt"

	appv1alpha1 "github.com/salkin/harbor-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_harbor")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Harbor Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHarbor{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("harbor-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Harbor
	err = c.Watch(&source.Kind{Type: &appv1alpha1.Harbor{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Harbor
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.Harbor{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileHarbor{}

// ReconcileHarbor reconciles a Harbor object
type ReconcileHarbor struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Harbor object and makes changes based on the state read
// and what is in the Harbor.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHarbor) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Harbor")

	// Fetch the Harbor instance
	instance := &appv1alpha1.Harbor{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and donj't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	config := r.verifyHarborConfig(instance)
	if config == nil {
		return reconcile.Result{}, fmt.Errorf("Failed to generate config")
	}

	//Verify postgresql
	err = r.verifyPostgre(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}

	config.HarborSecrets.DBUser = "postgres"
	config.HarborData.DBURL = instance.Name + "-database"
	config.HarborData.LogLevel = "info"

	err = r.verifyCore(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.verifyRegistry(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.verifyAdminserver(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.verifyJobService(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.verifyIngress(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.verifyPortal(instance, config)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileHarbor) verifyHarborConfig(cr *appv1alpha1.Harbor) *HarborInternal {
	cm := newCmForHarborInt(cr)
	e := &errCreator{c: r.verifyObject}
	e.create(cr, cm)

	intCm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name}, intCm)
	if err != nil {
		return nil
	}
	hb := &HarborInternal{}
	err = json.Unmarshal([]byte(cm.Data["data.json"]), hb)
	if err != nil {
		log.Error(err, "Failed to decode Harbor int")
	}
	return hb
}

func (r *ReconcileHarbor) verifyIngress(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}
	sec := newSecretForIngress(cr)
	e.create(cr, sec)

	ing := newIngressForCR(cr)
	e.create(cr, ing)
	if e.err != nil {
		return e.err
	}
	return nil

}

func (r *ReconcileHarbor) verifyJobService(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}
	sec := newSecretForJobservice(cr)
	e.create(cr, sec)

	if e.err != nil {
		return e.err
	}
	return nil
}
func (r *ReconcileHarbor) verifyPortal(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}
	dep := newPortalForCr(cr)
	e.create(cr, dep)

	svc := newServiceForPortal(cr)
	e.create(cr, svc)

	ing := newIngressForCR(cr)
	e.create(cr, ing)

	if e.err != nil {
		return e.err
	}
	return nil
}

func (r *ReconcileHarbor) verifyPostgre(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}
	//	ps := newPostgreCrd(cr)
	//	e.create(cr, ps)

	cm := newCmForDb(cr)
	e.create(cr, cm)

	sec := newSecretForDb(cr, d)
	e.create(cr, sec)

	ss := newStatefulSetForDb(cr)
	e.create(cr, ss)

	svc := newServiceForDb(cr)
	e.create(cr, svc)

	if e.err != nil {
		return e.err
	}
	return nil
}

func (r *ReconcileHarbor) verifyCore(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}
	cm := newCoreCmForCR(cr)
	e.create(cr, cm)

	svc := newServiceForCR(cr)
	e.create(cr, svc)

	secret := newSecretForCore(cr, d)
	e.create(cr, secret)

	depl := newCoreForCR(cr)
	e.create(cr, depl)
	if e.err != nil {
		return e.err
	}
	return nil
}

func (r *ReconcileHarbor) verifyRegistry(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}

	dep := newRegistryForCr(cr)
	e.create(cr, dep)

	svc := newServiceForRegistry(cr)
	e.create(cr, svc)

	sec := newSecretForRegistry(cr)
	e.create(cr, sec)

	cm := newCmForRegistry(cr, d, "")
	e.create(cr, cm)

	pvc := newPVCForRegistry(cr)
	e.create(cr, pvc)

	if e.err != nil {
		return e.err
	}
	return nil
}

func (r *ReconcileHarbor) verifyAdminserver(cr *appv1alpha1.Harbor, d *HarborInternal) error {
	e := &errCreator{c: r.verifyObject}
	dep := newAdminserverForCr(cr)
	e.create(cr, dep)

	sec := newSecretForAdminserver(cr, d)
	e.create(cr, sec)

	cm := newCmForAdminserver(cr, d, "")
	e.create(cr, cm)

	svc := newServiceForAdminserver(cr)
	e.create(cr, svc)
	if e.err != nil {
		return e.err
	}
	return nil
}

// verifyConfigMap retrieves configmap from API server and tries create if missing
func (r *ReconcileHarbor) verifyObject(cr *appv1alpha1.Harbor, cm runtime.Object) error {
	// Check if this Deployment already exists
	objKey, crErr := client.ObjectKeyFromObject(cm)
	if crErr != nil {
		return crErr
	}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: objKey.Name, Namespace: objKey.Namespace}, cm)
	return r.handleCreate(cr, err, cm)
}

func (r *ReconcileHarbor) handleCreate(cr *appv1alpha1.Harbor, err error, obj runtime.Object) error {
	if err != nil && errors.IsNotFound(err) {

		objKey, crErr := client.ObjectKeyFromObject(obj)
		if crErr != nil {
			return crErr
		}
		metaObj, err := meta.Accessor(obj)
		if err != nil {
			return err
		}
		r.setReference(cr, metaObj)
		log.Info("Creating a new  %s/%s\n", "Type", obj.GetObjectKind().GroupVersionKind().Kind, objKey.Namespace, objKey.Name)
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}
