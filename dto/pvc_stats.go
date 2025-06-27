package dto

// PVCStats represents processed statistics for a Kubernetes PVC resource
type PVCStats struct {
	Name             string   `json:"name"`
	Namespace        string   `json:"namespace"`
	Status           string   `json:"status"`
	StorageCapacity  string   `json:"storageCapacity"`
	StorageClassName string   `json:"storageClassName,omitempty"`
	AccessModes      []string `json:"accessModes"`
	VolumeName       string   `json:"volumeName,omitempty"`
	Created          string   `json:"created"`
	Phase            string   `json:"phase"`
	Annotations      []string `json:"annotations,omitempty"`
}

// PVCStatsResponse represents the response for PVC statistics API
type PVCStatsResponse struct {
	PVCs []PVCStats `json:"pvcs"`
}
