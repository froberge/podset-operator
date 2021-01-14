# Create the custom resources
echo "Create the CRD in the cluster"
kubectl apply -f ../deploy/crds/app.example.com_podsets_crd.yaml

# Create the RBAC necessary
#echo "Create the different RBAC required"
kubectl apply -f ../deploy/service_account.yaml
kubectl apply -f ../deploy/role.yaml
kubectl apply -f ../deploy/role_binding.yaml

# deploy the custom resources
# This command need to be activated if you want to deploy the operator into the cluster.
# To run the operator locally leave this comment
echo "Deploy the Operator"
kubectl apply -f ../deploy/operator.yaml
