package utils

import (
	corev1 "k8s.io/api/core/v1"
)

// GetNodeStatus determines the status of a node (Ready/NotReady)
func GetNodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return "Ready"
		}
	}
	return "NotReady"
}

// ExtractNodeRoles extracts roles from node labels
func ExtractNodeRoles(labels map[string]string) []string {
	roles := make([]string, 0)
	
	// Check for master/control-plane roles
	if _, ok := labels["node-role.kubernetes.io/master"]; ok {
		roles = append(roles, "master")
	}
	
	if _, ok := labels["node-role.kubernetes.io/control-plane"]; ok {
		roles = append(roles, "control-plane")
	}
	
	// Check for worker role
	if _, ok := labels["node-role.kubernetes.io/worker"]; ok {
		roles = append(roles, "worker")
	}
	
	// Check legacy role label
	if role, ok := labels["kubernetes.io/role"]; ok {
		roles = append(roles, role)
	}
	
	// If no roles were found, assume it's a worker
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	
	return roles
}
