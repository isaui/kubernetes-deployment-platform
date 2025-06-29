package utils

import (
	"context"
	"fmt"

	"github.com/pendeploy-simple/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// waitForRegistryPod waits for a pod to be created and returns its name
func WaitForRegistryPod(ctx context.Context, registry models.Registry, namespace string, clientset *kubernetes.Clientset) (string, error) {
	labelSelector := fmt.Sprintf("app=registry,registry-id=%s", registry.ID)
	
	// Poll until a pod is found or timeout
	var podName string
	err := retry.OnError(retry.DefaultRetry, func(error) bool { return true }, func() error {
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return err
		}
		
		if len(pods.Items) == 0 {
			return fmt.Errorf("pod not found")
		}
		
		// Get the first pod name
		podName = pods.Items[0].Name
		return nil
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to get registry pod: %w", err)
	}
	
	return podName, nil
}


