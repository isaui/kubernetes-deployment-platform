// utils/delete_kubernetes_resource_utils.go - UPDATE untuk handle managed services
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

// DeleteKubernetesResources deletes all Kubernetes resources for the service - UPDATED untuk managed services
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

	// Delete resources in reverse order for proper cleanup
	// Order: HPA -> Ingress -> Service -> Deployment/StatefulSet -> PVCs

	// Delete HPA if exists (both git and managed services may have HPA)
	err = k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Printf("Warning: Failed to delete HPA: %v", err)
	} else {
		log.Println("HPA deleted successfully")
	}

	// Delete Ingress
	err = k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Printf("Warning: Failed to delete Ingress: %v", err)
	} else {
		log.Println("Ingress deleted successfully")
	}

	// Delete Service
	err = k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service: %v", err)
	}
	log.Println("Service deleted successfully")

	// Delete based on service type
	if service.Type == models.ServiceTypeManaged {
		// For managed services, determine if it's StatefulSet or Deployment
		serviceType := GetManagedServiceType(service.ManagedType)
		
		if serviceType == "StatefulSet" {
			err = deleteStatefulSet(ctx, k8sClient, service)
			if err != nil {
				return fmt.Errorf("failed to delete StatefulSet: %v", err)
			}
			log.Println("StatefulSet deleted successfully")
			
			// Delete PVCs for StatefulSet (they are not auto-deleted)
			err = deleteManagedServicePVCs(ctx, k8sClient, service)
			if err != nil {
				log.Printf("Warning: Failed to delete PVCs: %v", err)
			} else {
				log.Println("PVCs deleted successfully")
			}
		} else {
			// Delete Deployment for managed services that use Deployment
			err = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("failed to delete Deployment: %v", err)
			}
			log.Println("Deployment deleted successfully")
			
			// Delete standalone PVC if exists
			err = deleteStandalonePVC(ctx, k8sClient, service)
			if err != nil {
				log.Printf("Warning: Failed to delete standalone PVC: %v", err)
			}
		}
	} else {
		// For git services, delete Deployment (original behavior)
		err = k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete Deployment: %v", err)
		}
		log.Println("Deployment deleted successfully")
	}

	return nil
}

// deleteStatefulSet deletes StatefulSet for managed services
func deleteStatefulSet(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	err := k8sClient.Clientset.AppsV1().StatefulSets(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete StatefulSet: %v", err)
	}
	
	log.Printf("StatefulSet %s deleted successfully", resourceName)
	return nil
}

// deleteManagedServicePVCs deletes PVCs created by StatefulSet volume claim templates
func deleteManagedServicePVCs(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	// StatefulSet creates PVCs with pattern: <volumeClaimTemplate>-<statefulset-name>-<ordinal>
	// Our volume claim template is named "data", so PVC will be "data-<resourceName>-0"
	pvcName := fmt.Sprintf("data-%s-0", resourceName)
	
	err := k8sClient.Clientset.CoreV1().PersistentVolumeClaims(service.EnvironmentID).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete PVC %s: %v", pvcName, err)
	}
	
	log.Printf("PVC %s deleted successfully", pvcName)
	return nil
}

// deleteStandalonePVC deletes standalone PVC for Deployment-based managed services
func deleteStandalonePVC(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	pvcName := fmt.Sprintf("%s-data", resourceName)
	
	err := k8sClient.Clientset.CoreV1().PersistentVolumeClaims(service.EnvironmentID).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Printf("Standalone PVC %s not found or already deleted", pvcName)
		return nil // Don't fail for missing standalone PVCs
	}
	
	log.Printf("Standalone PVC %s deleted successfully", pvcName)
	return nil
}