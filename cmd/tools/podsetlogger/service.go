package podsetlogger

import (
	appv1alpha1 "operator-framework/podset-operator/pkg/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//CreatePodsetloggerService generates the Podsetlogger service
func CreatePodsetloggerService(cr *appv1alpha1.PodSet) *corev1.Service {

	service := cr.Spec.PodSetLoggerService

	selectors := make(map[string]string, 0)
	for _, s := range cr.Spec.PodSetLoggerService.PodSelector {
		selectors[s.Name] = s.Value
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      service.ServiceName,
			Namespace: cr.Spec.Namespace,
		},

		Spec: corev1.ServiceSpec{
			Type:     service.ServiceType,
			Selector: selectors,
			Ports: []corev1.ServicePort{
				{
					Port: service.Ports.Port,
					TargetPort: intstr.IntOrString{
						IntVal: int32(service.Ports.TargetPort),
					},
				},
			},
		},
	}
}

// GetPodsetloggerService returns Podsetloggerservice
func GetPodsetloggerService(cr *appv1alpha1.PodSet) (*corev1.Service, *corev1.Service) {
	return &corev1.Service{}, CreatePodsetloggerService(cr)
}
