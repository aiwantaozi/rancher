package client

const (
	PersistentVolumeClaimSpecType                  = "persistentVolumeClaimSpec"
	PersistentVolumeClaimSpecFieldAccessModes      = "accessModes"
	PersistentVolumeClaimSpecFieldResources        = "resources"
	PersistentVolumeClaimSpecFieldSelector         = "selector"
	PersistentVolumeClaimSpecFieldStorageClassName = "storageClassName"
	PersistentVolumeClaimSpecFieldVolumeMode       = "volumeMode"
	PersistentVolumeClaimSpecFieldVolumeName       = "volumeName"
)

type PersistentVolumeClaimSpec struct {
	AccessModes      []string              `json:"accessModes,omitempty" yaml:"accessModes,omitempty"`
	Resources        *ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
	Selector         *LabelSelector        `json:"selector,omitempty" yaml:"selector,omitempty"`
	StorageClassName string                `json:"storageClassName,omitempty" yaml:"storageClassName,omitempty"`
	VolumeMode       string                `json:"volumeMode,omitempty" yaml:"volumeMode,omitempty"`
	VolumeName       string                `json:"volumeName,omitempty" yaml:"volumeName,omitempty"`
}
