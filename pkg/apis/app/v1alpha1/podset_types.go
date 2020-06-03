package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodSetSpec defines the desired state of PodSet
type PodSetSpec struct {
	Namespace           string              `json:"namespace"`
	PodSetLogger        Podsetlogger        `json:"podsetlogger-deployment-spec"`
	PodSetLoggerService PodSetloggerService `json:"podsetlogger-service-spec"`
	Watch               []Watch             `json:"watch"`

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// PodSetStatus defines the observed state of PodSet
type PodSetStatus struct {
	PodNames           []string   `json:"podnames,omitempty"`
	CurrentDeployment  Deployment `json:"currentdeployement,omitempty"`
	PreviousDeployment Deployment `json:"previousdeployement,omitempty"`
	Watch              []Watch    `json:"watch,omitempty"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// PodSetloggerService spec
type PodSetloggerService struct {
	ServiceName string             `json:"servicename"`
	ServiceType corev1.ServiceType `json:"servicetype"`
	PodSelector []Selectors        `json:"selectors"`
	Ports       Ports              `json:"ports"`
}

// Selectors spec
type Selectors struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Ports spec
type Ports struct {
	Port       int32 `json:"port"`
	TargetPort int32 `json:"targetport"`
}

// Podsetlogger contain the value necessary to make a deployment
type Podsetlogger struct {
	ImageName       string `json:"imagename"`
	Replicas        int32  `json:"replicas"`
	Version         string `json:"version"`
	ImageLocation   string `json:"imagelocation"`
	ImagePullPolicy string `json:"imagepullpolicy"`
}

// Deployment contain the value necessary to make a deployment
type Deployment struct {
	Name            string `json:"name"`
	Replicas        int32  `json:"replicas"`
	Version         string `json:"version"`
	ImageLocation   string `json:"imagelocation"`
	ImagePullPolicy string `json:"imagpullpolicy"`
	Err             string `json:"error"`
}

// Watch element
type Watch struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSet is the Schema for the podsets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=podsets,scope=Cluster
type PodSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodSetSpec   `json:"spec,omitempty"`
	Status PodSetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodSetList contains a list of PodSet
type PodSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodSet{}, &PodSetList{})
}
