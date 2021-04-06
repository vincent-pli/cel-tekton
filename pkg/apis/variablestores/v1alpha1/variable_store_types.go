/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// VariableStore is a context or variables storage to help caculate the CEL expression.
//
// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VariableStore struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the VariableStore (from the client).
	// +optional
	Spec VariableStoreSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the AddressableService (from the controller).
	// +optional
	Status VariableStoreStatus `json:"status,omitempty"`
}

var (
	// Check that VariableStore can be validated and defaulted.
	_ apis.Validatable   = (*VariableStore)(nil)
	_ apis.Defaultable   = (*VariableStore)(nil)
	_ kmeta.OwnerRefable = (*VariableStore)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*VariableStore)(nil)
)

// VariableStoreSpec holds the desired state of the VariableStore (from the client).
type VariableStoreSpec struct {
	// Vars holds the predefined variables and these variabls will be the context for next caculation.
	Vars []Var `json:"vars,omitempty"`
}

// Var declares an string to use for the var called name.
type Var struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

const (
	// AddressableServiceConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	AddressableServiceConditionReady = apis.ConditionReady
)

// VariableStoreStatus communicates the observed state of the AddressableService (from the controller).
type VariableStoreStatus struct {
	duckv1.Status `json:",inline"`
}

// AddressableServiceList is a list of AddressableService resources
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VariableStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VariableStore `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (vs *VariableStore) GetStatus() *duckv1.Status {
	return &vs.Status.Status
}
