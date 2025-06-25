package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// PodStatsService handles operations related to pod statistics
type PodStatsService struct {}

// NewPodStatsService creates a new pod stats service
func NewPodStatsService() *PodStatsService {
	return &PodStatsService{}
}

// GetPodStats returns statistics for pods in the specified namespace
// If namespace is empty, returns pods from all namespaces
func (s *PodStatsService) GetPodStats(namespace string) (dto.PodStatsResponse, error) {
	ctx := context.Background()
	log.Printf("Accessing pod metrics API for namespace %s", namespace)
	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return dto.PodStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	// Get pods from specified namespace (or all namespaces if empty)
	podList, err := kubeClient.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return dto.PodStatsResponse{}, fmt.Errorf("failed to list pods: %v", err)
	}
	
	// Get pod metrics if available
	var podMetricsList *metricsapi.PodMetricsList
	var metricsMap map[string]*metricsapi.PodMetrics
	
	if kubeClient.MetricsClient != nil {
		// Get pod metrics using the typed client
		log.Printf("Fetching pod metrics using typed client for namespace: %s", namespace)
		
		// Get pod metrics list
		var err error
		podMetricsList, err = kubeClient.MetricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
		
		if err != nil {
			// Log but continue - we can show pods without metrics
			log.Printf("Warning: Error getting pod metrics: %v\n", err)
		} else {
			log.Printf("Successfully retrieved metrics for %d pods", len(podMetricsList.Items))
			
			// Build metrics lookup map
			metricsMap = make(map[string]*metricsapi.PodMetrics, len(podMetricsList.Items))
			for i := range podMetricsList.Items {
				metrics := podMetricsList.Items[i]
				key := utils.GeneratePodKey(metrics.ObjectMeta.Namespace, metrics.ObjectMeta.Name)
				metricsMap[key] = &podMetricsList.Items[i]
			}
		}
	}

	// Process each pod
	podStats := make([]dto.PodStats, 0, len(podList.Items))
	for _, pod := range podList.Items {
		// Get pod type from owner references
		podType := utils.GetPodType(&pod)
		
		// Get container resources
		cpuRequest, cpuLimit, memoryRequest, memoryLimit := utils.GetContainerResourceTotals(&pod)
		
		// Format requests/limits
		cpuRequestStr := utils.FormatCPUCores(cpuRequest)
		cpuLimitStr := utils.FormatResourceValue(cpuLimit, utils.FormatCPUCores)
		memoryRequestStr := utils.FormatResourceValue(memoryRequest, utils.FormatBytesToHumanReadable)
		memoryLimitStr := utils.FormatResourceValue(memoryLimit, utils.FormatBytesToHumanReadable)

		// Initialize usage values
		cpuUsage := "0"
		memoryUsage := "0"
		var cpuPercentage float64 = 0
		var memoryPercentage float64 = 0

		// Get actual usage from metrics if available
		if metricsMap != nil {
			key := utils.GeneratePodKey(pod.Namespace, pod.Name)
			if metrics, ok := metricsMap[key]; ok {
				var totalCpuUsage int64 = 0
				var totalMemoryUsage int64 = 0

				// Sum usage across all containers
				for _, container := range metrics.Containers {
					totalCpuUsage += container.Usage.Cpu().MilliValue()
					totalMemoryUsage += container.Usage.Memory().Value()
				}

				// Format usage values
				cpuUsage = utils.FormatCPUCores(totalCpuUsage)
				memoryUsage = utils.FormatBytesToHumanReadable(totalMemoryUsage)

				// Calculate percentages
				if cpuLimit > 0 {
					cpuPercentage = utils.CalculatePercentage(totalCpuUsage, cpuLimit)
				} else if cpuRequest > 0 {
					// Use request if limit not set
					cpuPercentage = utils.CalculatePercentage(totalCpuUsage, cpuRequest)
				}

				if memoryLimit > 0 {
					memoryPercentage = utils.CalculatePercentage(totalMemoryUsage, memoryLimit)
				} else if memoryRequest > 0 {
					// Use request if limit not set
					memoryPercentage = utils.CalculatePercentage(totalMemoryUsage, memoryRequest)
				}
			}
		}

		// Add pod stats
		podStats = append(podStats, dto.PodStats{
			Name:           pod.Name,
			Namespace:      pod.Namespace,
			Status:         string(pod.Status.Phase),
			Type:           podType,
			ContainerCount: len(pod.Spec.Containers),
			CPU: dto.PodResource{
				Usage:      cpuUsage,
				Request:    cpuRequestStr,
				Limit:      cpuLimitStr,
				Percentage: cpuPercentage,
			},
			Memory: dto.PodResource{
				Usage:      memoryUsage,
				Request:    memoryRequestStr,
				Limit:      memoryLimitStr,
				Percentage: memoryPercentage,
			},
			Created: pod.CreationTimestamp.Format(time.RFC3339),
		})
	}

	// Return using proper DTO structure
	return dto.PodStatsResponse{
		Pods: podStats,
	}, nil
}
