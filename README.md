# Harbor-operator

Harbor-operator is a [Kubernetes operator]()

## Status

Alpha

### Quick start

To deploy a Harbor registry deploy the CRD and the operator 

```
kubectl create -f deploy/crds/app_v1alpha1_harbor_crd.yaml

kubectl create -f deploy
```

Apply an instance of the CRD to get a registry.
```
kubectl create -f deploy/crds/app_v1alpha1_harbor_cr.yaml
```
