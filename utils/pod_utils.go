package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// GetPodType determines a pod's type based on owner references
func GetPodType(pod *corev1.Pod) string {
	// Default pod type
	podType := "Pod"
	
	// Check owner references to determine pod type
	if len(pod.OwnerReferences) > 0 {
		ownerKind := pod.OwnerReferences[0].Kind
		switch ownerKind {
		case "ReplicaSet":
			podType = "Deployment" // Usually ReplicaSets are created by Deployments
		case "StatefulSet":
			podType = "StatefulSet"
		case "DaemonSet":
			podType = "DaemonSet"
		case "Job":
			podType = "Job"
		case "CronJob":
			podType = "CronJob"
		default:
			podType = ownerKind
		}
	}
	
	return podType
}

// GetContainerResourceTotals calculates the total resource requests and limits for a pod
// Returns CPU in milliCores and memory in bytes
func GetContainerResourceTotals(pod *corev1.Pod) (cpuRequest, cpuLimit, memoryRequest, memoryLimit int64) {
	for _, container := range pod.Spec.Containers {
		// Process CPU and memory requests
		if cpu := container.Resources.Requests.Cpu(); !cpu.IsZero() {
			cpuRequest += cpu.MilliValue()
		}
		
		if mem := container.Resources.Requests.Memory(); !mem.IsZero() {
			memoryRequest += mem.Value()
		}
		
		// Process CPU and memory limits
		if cpu := container.Resources.Limits.Cpu(); !cpu.IsZero() {
			cpuLimit += cpu.MilliValue()
		}
		
		if mem := container.Resources.Limits.Memory(); !mem.IsZero() {
			memoryLimit += mem.Value()
		}
	}
	
	return
}

// FormatResourceValue formats a resource value as string with proper units or infinity symbol if zero
func FormatResourceValue(value int64, formatter func(int64) string) string {
	if value > 0 {
		return formatter(value)
	}
	return "∞"  // Infinity symbol for unlimited resources
}

// GeneratePodKey generates a unique key for a pod based on namespace and name
func GeneratePodKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
