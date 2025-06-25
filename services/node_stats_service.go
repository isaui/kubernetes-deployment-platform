package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeletStorageData represents storage data structure from Kubelet API
type KubeletStorageData struct {
	UsedBytes     int64 `json:"usedBytes"`
	CapacityBytes int64 `json:"capacityBytes"`
}

// KubeletStatsSummary represents Kubelet stats summary response structure
type KubeletStatsSummary struct {
	Node struct {
		SystemContainers []struct{
			Name string `json:"name"`
			Fs   *KubeletStorageData `json:"fs,omitempty"`
		} `json:"systemContainers"`
		Fs *KubeletStorageData `json:"fs,omitempty"`
	} `json:"node"`
}

// NodeStatsService handles operations related to node statistics
type NodeStatsService struct {}

// NewNodeStatsService creates a new node stats service
func NewNodeStatsService() *NodeStatsService {
	return &NodeStatsService{}
}

// GetNodeStats returns statistics for nodes in the cluster
func (s *NodeStatsService) GetNodeStats() (dto.NodeStatsResponse, error) {
	ctx := context.Background()
	
	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return dto.NodeStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	// Get nodes information
	nodes, err := kubeClient.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return dto.NodeStatsResponse{}, fmt.Errorf("failed to list nodes: %v", err)
	}
	
	// Transform the nodes into our DTO format
	nodeStats := make([]dto.NodeStats, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		// Extract conditions
		conditions := make(dto.NodeConditions)
		for _, condition := range node.Status.Conditions {
			conditions[string(condition.Type)] = dto.NodeCondition{
				Status:            string(condition.Status),
				Reason:            condition.Reason,
				Message:           condition.Message,
				LastTransitionTime: condition.LastTransitionTime.String(),
			}
		}
		
		// Extract roles from labels
		roles := utils.ExtractNodeRoles(node.Labels)
		
		// Process CPU information
		cpuCapacity := node.Status.Capacity.Cpu().String()
		cpuAllocatable := node.Status.Allocatable.Cpu().String()
		
		// Process Memory information
		memoryCapacity := node.Status.Capacity.Memory().String()
		memoryAllocatable := node.Status.Allocatable.Memory().String()
		
		// Process Storage information
		storageCapacity := node.Status.Capacity.StorageEphemeral().String()
		storageAllocatable := node.Status.Allocatable.StorageEphemeral().String()

		// Default values
		var cpuUsage = "0"
		var memoryUsage = "0"
		var storageUsage = "0"
		var cpuPercentage = 0.0
		var memoryPercentage = 0.0
		var storagePercentage = 0.0

		// Find CPU and Memory usage from metrics if available
		if kubeClient.MetricsClient != nil {
			// Get node metrics from metrics API
			metrics, err := kubeClient.MetricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{})
			
			// If we can get metrics for this node
			if err == nil && metrics != nil {
				// CPU usage in cores
				cpuMilliCores := metrics.Usage.Cpu().MilliValue()
				cpuCores := float64(cpuMilliCores) / 1000.0
				cpuUsage = fmt.Sprintf("%.2f", cpuCores)
				
				// Memory usage in bytes, convert to readable format
				memoryBytes := metrics.Usage.Memory().Value()
				memoryUsage = utils.FormatBytesToHumanReadable(memoryBytes)
				
				// Try to get storage usage data from Kubelet API
				storageData, storageErr := s.getNodeStorageUsage(ctx, node.Name, kubeClient)
				if storageErr == nil && storageData != nil {
					// Format storage usage in readable format
					usedBytes := storageData.UsedBytes
					totalBytes := storageData.CapacityBytes
					
					storageUsage = utils.FormatBytesToHumanReadable(usedBytes)
					storageCapacity = utils.FormatBytesToHumanReadable(totalBytes)
					storageAllocatable = storageCapacity // Use consistent format
					
					// Calculate percentage used
					if totalBytes > 0 {
						storagePercentage = utils.CalculatePercentage(usedBytes, totalBytes)
					}
				}
				
				// Calculate CPU percentage
				cpuCapacityValue := node.Status.Capacity.Cpu().MilliValue()
				if cpuCapacityValue > 0 {
					cpuPercentage = utils.CalculatePercentage(cpuMilliCores, cpuCapacityValue)
				}
				
				// Calculate memory percentage
				memoryCapacityValue := node.Status.Capacity.Memory().Value()
				if memoryCapacityValue > 0 {
					memoryPercentage = utils.CalculatePercentage(memoryBytes, memoryCapacityValue)
				}
			}
		}
		
		// Add this node's stats to the result
		nodeStats = append(nodeStats, dto.NodeStats{
			Name:           node.Name,
			Status:         s.getNodeStatus(conditions),
			Conditions:     conditions,
			Roles:          roles,
			Created:        node.CreationTimestamp.String(),
			KubeletVersion: node.Status.NodeInfo.KubeletVersion,
			OSImage:        node.Status.NodeInfo.OSImage,
			CPU: dto.NodeResource{
				Capacity:    cpuCapacity,
				Allocatable: cpuAllocatable,
				Usage:       cpuUsage,
				Percentage:  cpuPercentage,
			},
			Memory: dto.NodeResource{
				Capacity:    memoryCapacity,
				Allocatable: memoryAllocatable,
				Usage:       memoryUsage,
				Percentage:  memoryPercentage,
			},
			Storage: dto.NodeResource{
				Capacity:    storageCapacity,
				Allocatable: storageAllocatable,
				Usage:       storageUsage,
				Percentage:  storagePercentage,
			},
		})
	}
	return dto.NodeStatsResponse{
		Nodes: nodeStats,
	}, nil
}

// getNodeStatus determines node status from conditions
func (s *NodeStatsService) getNodeStatus(conditions dto.NodeConditions) string {
	if ready, ok := conditions["Ready"]; ok {
		if ready.Status == "True" {
			return "Ready"
		}
	}
	
	// Check for specific not-ready reasons
	for _, cond := range conditions {
		if cond.Status != "True" && cond.Reason != "" {
			return fmt.Sprintf("NotReady: %s", cond.Reason)
		}
	}
	
	return "NotReady"
}

func (s *NodeStatsService) getNodeStorageUsage(ctx context.Context, nodeName string, kubeClient *kubernetes.Client) (*KubeletStorageData, error) {
	log.Printf("Accessing kubelet stats API for node %s", nodeName)
	
	// Pake Do() instead of DoRaw() buat dapet response object
	result := kubeClient.Clientset.CoreV1().RESTClient().Get().
		Resource("nodes").
		Name(nodeName).
		Suffix("proxy/stats/summary").
		Do(ctx)
	
	// Cek error
	if result.Error() != nil {
		log.Printf("Error accessing kubelet stats API: %v", result.Error())
		return nil, fmt.Errorf("error accessing kubelet stats API: %v", result.Error())
	}
	log.Printf("Accessing kubelet stats API for node %s success", nodeName)
	
	// Get status code
	var statusCode int
	result.StatusCode(&statusCode)
	log.Printf("Response status code: %d", statusCode)
	
	// Get raw data
	rawData, err := result.Raw()
	if err != nil {
		log.Printf("Error getting raw data: %v", err)
		return nil, fmt.Errorf("error getting raw data: %v", err)
	}
	

	
	// Cek kalo empty atau cuma whitespace
	if len(rawData) == 0 || len(strings.TrimSpace(string(rawData))) == 0 {
		return nil, fmt.Errorf("kubelet returned empty response for node %s", nodeName)
	}
	
	var summary KubeletStatsSummary
	if err := json.Unmarshal(rawData, &summary); err != nil {
		return nil, fmt.Errorf("error parsing kubelet stats: %v", err)
	}
	

	
	// Check if filesystem stats are available directly
	if summary.Node.Fs != nil && summary.Node.Fs.CapacityBytes > 0 {
		log.Printf("Found filesystem stats for node %s: %v", nodeName, summary.Node.Fs)
		return summary.Node.Fs, nil
	}
	
	// If not available directly, check system containers
	for _, container := range summary.Node.SystemContainers {
		log.Printf("Checking container: %s", container.Name)
		if container.Fs != nil {
			log.Printf("Container %s has filesystem stats: %v", container.Name, container.Fs)
		}
		
		if (container.Name == "kubelet" || container.Name == "runtime") && 
		   container.Fs != nil && container.Fs.CapacityBytes > 0 {
			log.Printf("Found filesystem stats for node %s: %v", nodeName, container.Fs)
			return container.Fs, nil
		}
	}
	
	return nil, fmt.Errorf("no storage stats available for node %s", nodeName)
}
