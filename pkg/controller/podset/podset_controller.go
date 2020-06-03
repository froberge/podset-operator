package podset

import (
	"context"
	"reflect"
	"time"

	tools "operator-framework/podset-operator/cmd/tools/podsetLogger"
	utils "operator-framework/podset-operator/cmd/utils"
	appv1alpha1 "operator-framework/podset-operator/pkg/apis/app/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_podset")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PodSet Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodSet{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconcilercar
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PodSet
	err = c.Watch(&source.Kind{Type: &appv1alpha1.PodSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner PodSet
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.PodSet{},
	})

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.PodSet{},
	})

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.PodSet{},
	})

	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePodSet implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePodSet{}

// ReconcilePodSet reconciles a PodSet object
type ReconcilePodSet struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PodSet object and makes changes based on the state read
// and what is in the PodSet.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePodSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PodSet")

	// Fetch the PodSet instance
	podSet := &appv1alpha1.PodSet{}

	err := r.client.Get(context.TODO(), request.NamespacedName, podSet)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// First make sure we have the Configuration for the Pod to work
	// Deal with the required ConfigMap
	existingPodsetConfigMap, podSetConfigMap := tools.GetConfigMap(podSet)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: podSetConfigMap.Name, Namespace: podSetConfigMap.Namespace}, existingPodsetConfigMap)

	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Create Config Map")
			err = r.client.Create(context.TODO(), podSetConfigMap)
			if err != nil {
				return reconcile.Result{}, err
			}

			podSet.Status.Watch = podSet.Spec.Watch
			r.client.Status().Update(context.TODO(), podSet)
		}

		t := time.Duration(10)
		return reconcile.Result{RequeueAfter: time.Second * t}, err
	}

	var isPreviousEqualSpec = tools.CompareDeploymentToSpec(&podSet.Spec.PodSetLogger, &podSet.Status.PreviousDeployment)

	if !tools.CompareDeploymentToSpec(&podSet.Spec.PodSetLogger, &podSet.Status.CurrentDeployment) &&
		(!isPreviousEqualSpec ||
			(isPreviousEqualSpec && podSet.Status.PreviousDeployment.Err == "")) {

		podSet.Status.PreviousDeployment = tools.CopyDeployment(&podSet.Status.CurrentDeployment)

		var cd = appv1alpha1.Deployment{
			Name:            podSet.Spec.PodSetLogger.ImageName,
			Replicas:        podSet.Spec.PodSetLogger.Replicas,
			Version:         podSet.Spec.PodSetLogger.Version,
			ImageLocation:   podSet.Spec.PodSetLogger.ImageLocation,
			ImagePullPolicy: podSet.Spec.PodSetLogger.ImagePullPolicy,
			Err:             "",
		}
		podSet.Status.CurrentDeployment = cd

		err := r.client.Status().Update(context.TODO(), podSet)
		if err != nil {
			reqLogger.Error(err, "failed to update the podSet")
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	}

	if len(podSet.Status.CurrentDeployment.Version) == 0 {
		return reconcile.Result{}, nil
	}

	// List all the pods owned by this PodSet instance at any level of version
	labelsSet := labels.Set{
		"app": podSet.Status.CurrentDeployment.Name,
	}

	existingPods := &corev1.PodList{}

	// Find all the pods that match the labelsSet
	err = r.client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     podSet.Spec.Namespace,
			LabelSelector: labels.SelectorFromSet(labelsSet),
		})
	if err != nil {
		reqLogger.Error(err, "failed to list existing pods in the podSet")
		return reconcile.Result{}, err
	}

	existingPodNames := []string{}

	for _, pod := range existingPods.Items {
		// This pod is mark for deletetion, forget
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}

		// Check the version running, if not right version delete the pod
		if pod.GetLabels()["version"] != podSet.Status.CurrentDeployment.Version {
			err := deletePod(r, &reqLogger, &pod)
			if err != nil {
				reqLogger.Error(err, "failed to delete a pod with previous version")
				return reconcile.Result{}, err
			}
			continue
		} else if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {

			// Check the state of the container to make sure it in a healthy state.
			if len(pod.Status.ContainerStatuses) > 0 {
				containerStatus := pod.Status.ContainerStatuses[0]

				if !containerStatus.Ready {
					if containerStatus.State.Waiting != nil {
						if containerStatus.State.Waiting.Reason != utils.ContainerCreating {
							reqLogger.Info("ROLLBACK to previous version", "podName", pod.GetName())

							// Keep the version of the previous deployment
							// Rollback the version
							tmpDeployment := tools.CopyDeployment(&podSet.Status.PreviousDeployment)

							// The previous deployment because the current deployment with and err
							previousDeployment := tools.CopyDeployment(&podSet.Status.CurrentDeployment)
							previousDeployment.Err = containerStatus.State.Waiting.Reason
							podSet.Status.PreviousDeployment = previousDeployment

							// Update the state of the deployment
							podSet.Status.CurrentDeployment = tmpDeployment

							r.client.Status().Update(context.TODO(), podSet)
							return reconcile.Result{}, nil
						}
					}
				}
			}

			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	reqLogger.Info("Checking podset - ", "expected replicas", podSet.Status.CurrentDeployment.Replicas, "Pod.Names", existingPodNames)

	// Scale down number of replicas if to many node.
	if int32(len(existingPodNames)) > podSet.Status.CurrentDeployment.Replicas {
		// Delete a pod since their is to many
		reqLogger.Info("Deleting a pod in the podset", "expecting replicas", podSet.Status.CurrentDeployment.Replicas, "Pod.Names", existingPodNames)
		err = deletePod(r, &reqLogger, &existingPods.Items[0])

		if err != nil {
			reqLogger.Error(err, "failed to delete a pod with previous version")
			return reconcile.Result{}, err
		}
	}

	// Scale Up Pods
	if int32(len(existingPodNames)) < podSet.Status.CurrentDeployment.Replicas {
		// create a new pod & Set the PodSet as the owner and controller.
		reqLogger.Info("Adding a pod in the podset", "expected replicas", podSet.Status.CurrentDeployment.Replicas, "Pod.Names", existingPodNames)
		pod := tools.CreatePodsetloggerDeployment(&podSet.Status.CurrentDeployment, podSet.Spec.Namespace)
		if err := controllerutil.SetControllerReference(podSet, pod, r.scheme); err != nil {
			reqLogger.Error(err, "unable to set owner reference on new pod")
			return reconcile.Result{}, err
		}

		// Create the Pod
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			reqLogger.Error(err, "failed to create a pod")
			return reconcile.Result{}, err
		}

		existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
	}

	if !reflect.DeepEqual(podSet.Status.PodNames, existingPodNames) {
		podSet.Status.PodNames = existingPodNames
		err := r.client.Status().Update(context.TODO(), podSet)
		if err != nil {
			reqLogger.Error(err, "failed to update the podSet")
			return reconcile.Result{}, err
		}
	}

	// Create the require service
	reqLogger.Info(" ** CREATE THE REQUIRE SERVICE IF NOT FOUND ** ")
	existingPodsetloggerService, podSetLoggerService := tools.GetPodsetloggerService(podSet)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: podSetLoggerService.Name, Namespace: podSetLoggerService.Namespace}, existingPodsetloggerService)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Create Service")
			err = r.client.Create(context.TODO(), tools.CreatePodsetloggerService(podSet))
			if err != nil {
				return reconcile.Result{}, nil
			}
		}

		// Requeue, but wait 5 second to hage time to create the service
		t := time.Duration(5)
		return reconcile.Result{RequeueAfter: time.Second * t}, err
	}

	// Check if the config map need to be updated
	reqLogger.Info(" ** CHECK IF CONFIG NEED TO BE UPDATED ** ")
	eq := reflect.DeepEqual(podSet.Spec.Watch, podSet.Status.Watch)

	if !eq {
		reqLogger.Info("Update the Config Map")
		existingPodsetConfigMap.Data = podSetConfigMap.Data
		err := r.client.Update(context.TODO(), existingPodsetConfigMap)
		if err != nil {
			reqLogger.Error(err, "failed to update the podSet")
		} else {
			reqLogger.Info("Config Map was updated")
			podSet.Status.Watch = podSet.Spec.Watch
			r.client.Status().Update(context.TODO(), podSet)

			return reconcile.Result{Requeue: true}, nil
		}
	}

	return reconcile.Result{Requeue: true}, nil
}

// Delete a given pod
func deletePod(r *ReconcilePodSet, reqLogger *logr.Logger, pod *corev1.Pod) error {
	(*reqLogger).Info("Deleting a pod", "Pod.Version", pod.GetLabels()["version"], "Pod.Name", pod.GetName())
	return r.client.Delete(context.TODO(), pod)
}
