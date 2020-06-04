# podset-operator
This is my first try at creating a GO operator for Kubernetes using the [operator-sdk](https://sdk.operatorframework.io/)

## Prerequisites

1.  Install the [*operator sdk for GO*](https://sdk.operatorframework.io/docs/golang/installation/) with the CLI and it required tools.
2. Have access tp a Kubenetes Cluster. *For this example I've used Google Cloud Platform.  You can find a script to create a Cluster in the script folder in this repository.  Please adapt to relect your reality*

This operator is watching 3 types of ressources.

1. The Deployment
    * It check that the right version is deployed.  If the deployment fail it will rollback to the previous version.  It also make sure that the number of replicas is respected.
2. The Services
    * It will create the service to expose the applicaiton to the external world if the service doesnt' exist.
3. The Configuration Map
    * It will create a configuration map that is expose in a Volume.  If the configuration change in the **CR** it will update the configuration Map.,  It can take up to 1 minutes for the information to be replicated in the application.


## How to Run


