package utils

import (
	"context"
	"fmt"
	"log"
	
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// DeleteKubernetesResources deletes all Kubernetes resources for the service
func DeleteKubernetesResources(service models.Service) error {
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
    log.Println("Kubernetes client created successfully")
	// Create context for the operations
	ctx := context.Background()
	
	// Get consistent resource name
	resourceName := GetResourceName(service)

	// Delete resources in reverse order (HPA -> Ingress -> Service -> Deployment)
	// This ensures proper cleanup

	// Delete HPA if exists
	err = k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete HPA: %v", err)
	}
    log.Println("HPA deleted successfully")
	// Delete Ingress
	err = k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Ingress: %v", err)
	}
    log.Println("Ingress deleted successfully")
	// Delete Service
	err = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service: %v", err)
	}

	// Delete Deployment
	err = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment: %v", err)
	}

	return nil
}