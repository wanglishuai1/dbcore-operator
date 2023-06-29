package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DbConfig
type DbConfig struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DbConfigSpec `json:"spec"`
	// +optional
	Status DbConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DbConfigs
type DbConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DbConfig `json:"items"`
}

type DbConfigSpec struct {
	Replicas    int    `json:"replicas,omitempty"`
	Dsn         string `json:"dsn,omitempty"`
	MaxOpenConn int    `json:"maxOpenConn,omitempty"`
	MaxLifeTime int    `json:"maxLifeTime,omitempty"`
	MaxIdleConn int    `json:"maxIdleConn,omitempty"`
}

type DbConfigStatus struct {
	Replicas int32  `json:"replicas,omitempty"`
	Ready    string `json:"ready,omitempty"`
}
