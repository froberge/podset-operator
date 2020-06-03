package podsetlogger

import (
	"math/rand"
	"strconv"

	appv1alpha1 "operator-framework/podset-operator/pkg/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreatePodsetloggerDeployment creates a new deployment for the Podsetlogger
func CreatePodsetloggerDeployment(cr *appv1alpha1.Deployment, namespace string) *corev1.Pod {
	labels := map[string]string{
		"app":     cr.Name,
		"version": cr.Version,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-" + cr.Version + "-pod" + strconv.Itoa(rand.Intn(100)),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            cr.Name,
					Image:           cr.ImageLocation + cr.Name + ":" + cr.Version,
					ImagePullPolicy: corev1.PullPolicy(cr.ImagePullPolicy),
					VolumeMounts:    generateVolumeMounts(),
				},
			},
			Volumes: generateVolumes(),
		},
	}
}

//CompareDeploymentToSpec compare if a given deployment is equal to the on define in the spec
func CompareDeploymentToSpec(spec *appv1alpha1.Podsetlogger, deployment *appv1alpha1.Deployment) bool {
	return spec.ImageName == deployment.Name &&
		spec.Version == deployment.Version &&
		spec.Replicas == deployment.Replicas &&
		spec.ImageLocation == deployment.ImageLocation &&
		spec.ImagePullPolicy == deployment.ImagePullPolicy
}

// CopyDeployment is a way to recreate a new deployment with the same values
func CopyDeployment(deployment *appv1alpha1.Deployment) appv1alpha1.Deployment {
	var newDeployment = appv1alpha1.Deployment{
		Name:            deployment.Name,
		Replicas:        deployment.Replicas,
		Version:         deployment.Version,
		ImageLocation:   deployment.ImageLocation,
		ImagePullPolicy: deployment.ImagePullPolicy,
		Err:             deployment.Err,
	}

	return newDeployment
}
