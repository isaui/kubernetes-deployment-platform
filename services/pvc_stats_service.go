package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVCStatsService handles operations related to Kubernetes PVC resources statistics
type PVCStatsService struct{}

// NewPVCStatsService creates a new PVC stats service
func NewPVCStatsService() *PVCStatsService {
	return &PVCStatsService{}
}

// GetPVCStats returns statistics about PVC resources in the specified namespace
func (s *PVCStatsService) GetPVCStats(namespace string) (dto.PVCStatsResponse, error) {
	ctx := context.Background()

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return dto.PVCStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get PVC resources
	pvcList, err := kubeClient.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return dto.PVCStatsResponse{}, fmt.Errorf("failed to list PVCs: %v", err)
	}

	// Process each PVC
	pvcStats := make([]dto.PVCStats, 0, len(pvcList.Items))
	for _, pvc := range pvcList.Items {
		// Extract access modes
		accessModes := make([]string, 0, len(pvc.Spec.AccessModes))
		for _, mode := range pvc.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		// Extract storage capacity
		capacity := "Unknown"
		if pvc.Spec.Resources.Requests != nil {
			if storage, ok := pvc.Spec.Resources.Requests["storage"]; ok {
				capacity = storage.String()
			}
		}

		// Extract storage class name
		storageClassName := ""
		if pvc.Spec.StorageClassName != nil {
			storageClassName = *pvc.Spec.StorageClassName
		}

		// Extract notable annotations
		annotations := make([]string, 0)
		for k, v := range pvc.Annotations {
			annotations = append(annotations, fmt.Sprintf("%s: %s", k, v))
		}

		// Create PVCStats
		pvcStats = append(pvcStats, dto.PVCStats{
			Name:             pvc.Name,
			Namespace:        pvc.Namespace,
			Status:           string(pvc.Status.Phase),
			StorageCapacity:  capacity,
			StorageClassName: storageClassName,
			AccessModes:      accessModes,
			VolumeName:       pvc.Spec.VolumeName,
			Created:          pvc.CreationTimestamp.Format(time.RFC3339),
			Phase:            string(pvc.Status.Phase),
			Annotations:      annotations,
		})
	}

	return dto.PVCStatsResponse{
		PVCs: pvcStats,
	}, nil
}
