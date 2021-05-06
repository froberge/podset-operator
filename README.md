# PodSet Operator
This is a POC for k8s operator using Go and the [operator-sdk](https://sdk.operatorframework.io/)

## Prerequisites

1.  Install the [*operator sdk for GO*](https://sdk.operatorframework.io/docs/golang/installation/) with the CLI and it required tools.
2. You need access to a Kubenetes Cluster. 
3. I suggest using [Visual Studio Code](https://code.visualstudio.com/).
4. Import the following [GitHub project - PodSet Logger ](https://github.com/froberge-cloudOps/podSetLogger) to create the require docker images.

## Deployment

The operator can be deployed in two different ways. Locally and inside a Kubenetes Cluster.  In both scenario you need access to a Kubenetes Cluster. You will have to confirgure the Cluster the same way in both case. The only difference is where the operator is running.

#### Configuring the Cluster
1. Apply the Custom Resource Definition in your cluster.  This is done by running the *.crd.yaml file.
    >  kubectl apply -f ../deploy/crds/app.example.com_podsets_crd.yaml
2. Apply the different RBAC file in your cluster.

    > kubectl apply -f ../deploy/service_account.yaml

    > kubectl apply -f ../deploy/role.yaml

    > kubectl apply -f ../deploy/role_binding.yaml


#### Running Locally

* You can run the operator locally by running the following command.
    > operator-sdk run --local --namespace=podset-operator-group

* You can also run it in debug mode using [delve](https://github.com/go-delve/delve) which is a degugger for Go applicaiton.

    > operator-sdk run --local --namespace=podset-operator-group --enable-delve

#### Running in the cluster
* To run in the cluster you need to deploy the operator by running the deploy/opertor.yaml file with the following command.
    > kubectl apply -f deploy/operator.yamlkubectl apply -f deploy/operator.yaml

## The Operator
This operator is watching 3 types of ressources.

1. The Deployment
    * It check that the right version is deployed.  If the deployment fail it will rollback to the previous version.  It also make sure that the number of replicas is respected.
2. The Services
    * It will create the service to expose the applicaiton to the external world if the service doesnt' exist.
3. The Configuration Map
    * It will create a configuration map that is expose in a Volume.  If the configuration change in the **CR** it will update the configuration Map.,  It can take up to 1 minutes for the information to be replicated in the application.



## Authors
[Felix Roberge](https://github.com/roberge.felix@gmail.com) 
