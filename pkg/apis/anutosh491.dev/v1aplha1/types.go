package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterID",type=string,JSONPath=`.status.Digital_Ocean_ClusterID`
// +kubebuilder:printcolumn:name="Progress",type=string,JSONPath=`.status.progress`
type Digital_Ocean_Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Digital_Ocean_ClusterSpec   `json:"spec,omitempty"`
	Status Digital_Ocean_ClusterStatus `json:"status,omitempty"`
}

type KlsuterStatus struct {
	Digital_Ocean_ClusterID  string `json:"Digital_Ocean_ClusterID,omitempty"`
	Progress   string `json:"progress,omitempty"`
	KubeConfig string `json:"kubeConfig,omitempty"`
}

type Digital_Ocean_ClusterSpec struct {
	Name        string `json:"name,omitempty"`
	Region      string `json:"region,omitempty"`
	Version     string `json:"version,omitempty"`
	TokenSecret string `json:"tokenSecret,omitempty"`

	NodePools []NodePool `json:"nodePools,omitempty"`
}

type NodePool struct {
	Size  string `json:"size,omitempty"`
	Name  string `json:"name,omitempty"`
	Count int    `json:"count,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Digital_Ocean_ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Digital_Ocean_Cluster `json:"items,omitempty"`
}
