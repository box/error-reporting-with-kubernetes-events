package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: For a more complete example, this file could be the source of code
// autogeneration. (Similar to simple controller
// https://github.com/kubernetes/sample-controller/blob/master/pkg/apis/samplecontroller/v1alpha1/types.go#L23)
// However the purpose of this example is to demonstrate Kubernetes Events usage
// for error messages. We have not used auto code generation for that.

// Pki is the specification of a pki.box.com resource
type Pki struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PkiSpec `json:"spec"`
}

// PkiSpec is the spec for a Pki resource. For the sake of simplicity it
// only contains one field.
type PkiSpec struct {
	ServiceName string `json:"serviceName"`
}
