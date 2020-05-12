package podset

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"

	"operator-framework/podset-operator/pkg/apis/app/v1alpha1"
	appv1alpha1 "operator-framework/podset-operator/pkg/apis/app/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PodSet
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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

type deployment struct {
	name          string
	replicas      int32
	version       string
	imageLocation string
	pullPolicy    string
	err           bool
}

// DeployedState define the status of a deployment
type deployedState struct {
	state   string
	message string
}

var previousDeployment deployment
var currentDeployment deployment

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

	var isPreviousEqualSpec = compareDeploymentToSpec(podSet, previousDeployment)

	if !compareDeploymentToSpec(podSet, currentDeployment) &&
		(!isPreviousEqualSpec ||
			(isPreviousEqualSpec && previousDeployment.err == false)) {

		previousDeployment = currentDeployment
		currentDeployment = deployment{
			name:          podSet.Name,
			replicas:      podSet.Spec.Replicas,
			version:       podSet.Spec.Version,
			imageLocation: podSet.Spec.ImageLocation,
			pullPolicy:    podSet.Spec.ImagePullPolicy,
			err:           false,
		}

		if len(podSet.Status.DeploymentStatus) > 0 {
			podSet.Status.DeploymentStatus = ""

			err := r.client.Status().Update(context.TODO(), podSet)
			if err != nil {
				reqLogger.Error(err, "failed to update the podSet")
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{Requeue: true}, nil
	}

	if len(currentDeployment.version) == 0 {
		return reconcile.Result{}, nil
	}

	// List all the pods owned by this PodSet instance at any level of version
	labelsSet := labels.Set{
		"app": podSet.Name,
	}

	existingPods := &corev1.PodList{}

	// Find all the pods that match the labelsSet
	err = r.client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     request.Namespace,
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
		if pod.GetLabels()["version"] != currentDeployment.version {
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
						if containerStatus.State.Waiting.Reason != appv1alpha1.ContainerCreating {
							reqLogger.Info("ROLLBACK to previous version", "podName", pod.GetName())

							// Keep the version of the previous deployment
							// Rollback the version
							tmpDep := deployment{
								name:          previousDeployment.name,
								replicas:      previousDeployment.replicas,
								version:       previousDeployment.version,
								imageLocation: previousDeployment.imageLocation,
								pullPolicy:    previousDeployment.pullPolicy,
								err:           previousDeployment.err,
							}

							// The previous deployment because the current deployment with and err
							previousDeployment = currentDeployment
							previousDeployment.err = true

							// Current deployment become the previous deployment that was successfull.
							currentDeployment = tmpDep

							// Update the state of the deployment
							podSet.Status.DeploymentStatus = "ROLLBACK"
							r.client.Status().Update(context.TODO(), podSet)
							return reconcile.Result{}, nil
						}
					}
				}
			}

			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	reqLogger.Info("Checking podset - ", "expected replicas", currentDeployment.replicas, "Pod.Names", existingPodNames)

	// Scale down number of replicas if to many node.
	if int32(len(existingPodNames)) > currentDeployment.replicas {
		// Delete a pod since their is to many
		reqLogger.Info("Deleting a pod in the podset", "expecting replicas", currentDeployment.replicas, "Pod.Names", existingPodNames)
		err = deletePod(r, &reqLogger, &existingPods.Items[0])

		if err != nil {
			reqLogger.Error(err, "failed to delete a pod with previous version")
			return reconcile.Result{}, err
		}
	}

	// Scale Up Pods
	if int32(len(existingPodNames)) < currentDeployment.replicas {
		// create a new pod & Set the PodSet as the owner and controller.
		reqLogger.Info("Adding a pod in the podset", "expected replicas", currentDeployment.replicas, "Pod.Names", existingPodNames)
		pod := createNewPod(currentDeployment, podSet.Namespace)
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

	return reconcile.Result{Requeue: true}, nil
}

// Instantiate a new pod with the proper information.
func createNewPod(currentDeployment deployment, namespace string) *corev1.Pod {

	labels := map[string]string{
		"app":     currentDeployment.name,
		"version": currentDeployment.version,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      currentDeployment.name + "-" + currentDeployment.version + "-pod" + strconv.Itoa(rand.Intn(100)),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            currentDeployment.name,
					Image:           currentDeployment.imageLocation + currentDeployment.name + ":" + currentDeployment.version,
					ImagePullPolicy: corev1.PullPolicy(currentDeployment.pullPolicy),
				},
			},
		},
	}
}

func deletePod(r *ReconcilePodSet, reqLogger *logr.Logger, pod *corev1.Pod) error {
	(*reqLogger).Info("Deleting a pod", "Pod.Version", pod.GetLabels()["version"], "Pod.Name", pod.GetName())
	return r.client.Delete(context.TODO(), pod)
}

// Compare if two deployment are the same
func compareDeploymentToSpec(podSet *v1alpha1.PodSet, deployment deployment) bool {

	if podSet.Name != deployment.name {
		return false
	}

	return podSet.Spec.Version == deployment.version &&
		podSet.Spec.Replicas == deployment.replicas &&
		podSet.Spec.ImageLocation == deployment.imageLocation &&
		podSet.Spec.ImagePullPolicy == deployment.pullPolicy
}
