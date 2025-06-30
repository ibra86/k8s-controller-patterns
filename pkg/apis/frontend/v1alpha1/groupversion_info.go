// +kubebuilder:object:generate=true
// +groupName=frontendpage.ibra86.io

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "frontendpage.ibra86.io", Version: "v1alpha1"}

	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)
