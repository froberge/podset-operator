# Build the operator images & push it to the registry
#echo "Building and push the operator image"
#op-sdk build gcr.io/tools-poc/podset-operator
#docker push gcr.io/tools-poc/podset-operator

# Create the GCP cluster
echo "Create the GCP cluster"
gcloud container clusters create test-operator-fr --zone northamerica-northeast1-a

# Create the require namespace
echo "Create the required namespace"
kubectl apply -f deploy/k8s/namespace.yaml

# Create the custom resources on GCP
echo "Create the CRD in the cluster"
kubectl apply -f deploy/crds/app.example.com_podsets_crd.yaml

# Create the RBAC necessary
echo "Create the different RBAC required"
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml

# deploy the custom resources
#echo "Deploythe CRD"
kubectl apply -f deploy/operator.yaml
