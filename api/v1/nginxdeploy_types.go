/*


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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NginxDeploySpec defines the desired state of NginxDeploy
type NginxDeploySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NginxDeploy. Edit NginxDeploy_types.go to remove/update
	Image string `json:"image,omitempty"`
	Replicas  int32 `json:"replicas,omitempty"`
}

// NginxDeployStatus defines the observed state of NginxDeploy
type NginxDeployStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// NginxDeploy is the Schema for the nginxdeploys API
type NginxDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxDeploySpec   `json:"spec,omitempty"`
	Status NginxDeployStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NginxDeployList contains a list of NginxDeploy
type NginxDeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxDeploy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NginxDeploy{}, &NginxDeployList{})
}
