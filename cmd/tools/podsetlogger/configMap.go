package podsetlogger

import (
	"bytes"
	"fmt"
	appv1alpha1 "operator-framework/podset-operator/pkg/apis/app/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateVolumeMounts value.
func generateVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "podset-logger-config",
			MountPath: "/etc/config",
		},
	}
}

// GenerateVolumeMount
func generateVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "podset-logger-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "podset-logger-config",
					},
				},
			},
		},
	}
}

// GetConfigMap returns the podSet config map
func GetConfigMap(cr *appv1alpha1.PodSet) (*corev1.ConfigMap, *corev1.ConfigMap) {
	return &corev1.ConfigMap{}, CreateConfigMap(cr)
}

// CreateConfigMap - generating a conffig map
func CreateConfigMap(cr *appv1alpha1.PodSet) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podset-logger-config",
			Namespace: cr.Spec.Namespace,
		},
		Data: map[string]string{
			"podset-logger-config.yaml": generateConfig(cr),
		},
	}
}

func generateConfig(cr *appv1alpha1.PodSet) string {
	podSetLogger := make(map[string]string, 0)
	for _, w := range cr.Spec.Watch {
		podSetLogger[w.Name] = w.Value
	}

	output := new(bytes.Buffer)
	for key, value := range podSetLogger {
		fmt.Fprintf(output, "%s: \"%s\"\n", key, value)
	}

	return output.String()
}
