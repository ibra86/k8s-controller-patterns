package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FrontendPageSpec struct {
	Contents string `json:"contents"`
	Image    string `json:"image"`
	Replicas int    `json:"replicas"`
}

// +kubebuilder:object:root=true
type FrontendPage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FrontendPageSpec `json:"spec"`
}

// +kubebuilder:object:root=true
type FrontendPageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrontendPage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrontendPage{}, &FrontendPageList{})
}
