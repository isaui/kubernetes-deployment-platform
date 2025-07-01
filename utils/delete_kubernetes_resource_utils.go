// utils/delete_kubernetes_resource_utils.go
package utils

import (
	"context"
	"fmt"
	"log"
	
	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeleteKubernetesResources deletes all Kubernetes resources for the service with full managed service support
func DeleteKubernetesResources(service models.Service) error {
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	log.Println("Kubernetes client created successfully")
	
	// Create context for the operations
	ctx := context.Background()

	// Delete resources in reverse order for proper cleanup
	// Order: HPA -> All Ingresses & TCP Routes -> All Services -> Deployment/StatefulSet -> PVCs

	// Delete HPA if exists (both git and managed services may have HPA)
	if err := deleteHPAData(ctx, k8sClient, service); err != nil {
		log.Printf("Warning: Failed to delete HPA: %v", err)
	}

	// Delete all Ingresses and TCP Routes (managed services may have multiple)
	if err := deleteAllIngressesAndTCPRoutes(ctx, k8sClient, service); err != nil {
		log.Printf("Warning: Failed to delete all Ingresses and TCP Routes: %v", err)
	}

	// Delete all Services (managed services may have multiple)
	if err := deleteAllServices(ctx, k8sClient, service); err != nil {
		return fmt.Errorf("failed to delete Services: %v", err)
	}

	// Delete workload based on service type
	if service.Type == models.ServiceTypeManaged {
		if err := deleteManagedServiceWorkload(ctx, k8sClient, service); err != nil {
			return fmt.Errorf("failed to delete managed service workload: %v", err)
		}
		
		// Delete all PVCs for managed services
		if err := deleteAllManagedServicePVCs(ctx, k8sClient, service); err != nil {
			log.Printf("Warning: Failed to delete all PVCs: %v", err)
		}
	} else {
		// For git services, delete Deployment (original behavior)
		if err := deleteDeployment(ctx, k8sClient, service); err != nil {
			return fmt.Errorf("failed to delete Deployment: %v", err)
		}
	}

	log.Printf("Successfully deleted all resources for service: %s", service.Name)
	return nil
}

// deleteHPA deletes HorizontalPodAutoscaler
func deleteHPAData(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	err := k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	
	if err == nil {
		log.Printf("HPA %s deleted successfully", resourceName)
	}
	return nil
}

// deleteAllIngressesAndTCPRoutes deletes all ingresses and TCP routes for a service
func deleteAllIngressesAndTCPRoutes(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	if service.Type == models.ServiceTypeManaged {
		// Delete all ingresses and TCP routes for managed service based on exposure config
		exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)
		
		for _, config := range exposureConfigs {
			routeName := resourceName
			if config.Name != "primary" {
				routeName = fmt.Sprintf("%s-%s", resourceName, config.Name)
			}
			
			if config.IsHTTP {
				// Delete HTTP Ingress
				err := k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Delete(ctx, routeName, metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					log.Printf("Warning: Failed to delete HTTP Ingress %s: %v", routeName, err)
				} else if err == nil {
					log.Printf("HTTP Ingress %s deleted successfully", routeName)
				}
			} else {
				// Delete TCP IngressRoute
				if err := deleteTCPIngressRoute(ctx, k8sClient, service.EnvironmentID, routeName); err != nil {
					log.Printf("Warning: Failed to delete TCP IngressRoute %s: %v", routeName, err)
				} else {
					log.Printf("TCP IngressRoute %s deleted successfully", routeName)
				}
			}
		}
	} else {
		// Delete single ingress for git services
		err := k8sClient.Clientset.NetworkingV1().Ingresses(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if err == nil {
			log.Printf("Ingress %s deleted successfully", resourceName)
		}
	}
	
	return nil
}

// deleteTCPIngressRoute deletes Traefik IngressRouteTCP
func deleteTCPIngressRoute(ctx context.Context, k8sClient *kubernetes.Client, namespace, name string) error {
	gvr := schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutetcps",
	}

	dynamicClient := k8sClient.DynamicClient.Resource(gvr).Namespace(namespace)
	
	err := dynamicClient.Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	
	return nil
}

// deleteAllServices deletes all services for a service (multiple for managed services)
func deleteAllServices(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	if service.Type == models.ServiceTypeManaged {
		// Delete all services for managed service based on exposure config
		exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)
		
		for _, config := range exposureConfigs {
			serviceName := resourceName
			if config.Name != "primary" {
				serviceName = fmt.Sprintf("%s-%s", resourceName, config.Name)
			}
			
			err := k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, serviceName, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("failed to delete Service %s: %v", serviceName, err)
			}
			if err == nil {
				log.Printf("Service %s deleted successfully", serviceName)
			}
		}
	} else {
		// Delete single service for git services
		err := k8sClient.Clientset.CoreV1().Services(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete Service %s: %v", resourceName, err)
		}
		if err == nil {
			log.Printf("Service %s deleted successfully", resourceName)
		}
	}
	
	return nil
}

// deleteManagedServiceWorkload deletes the main workload for managed services (always StatefulSet)
func deleteManagedServiceWorkload(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	// All managed services use StatefulSet
	return deleteStatefulSet(ctx, k8sClient, service)
}

// deleteStatefulSet deletes StatefulSet for managed services
func deleteStatefulSet(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	err := k8sClient.Clientset.AppsV1().StatefulSets(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	
	if err == nil {
		log.Printf("StatefulSet %s deleted successfully", resourceName)
	}
	return nil
}

// deleteDeployment deletes Deployment for both git and managed services
func deleteDeployment(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	err := k8sClient.Clientset.AppsV1().Deployments(service.EnvironmentID).Delete(ctx, resourceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	
	if err == nil {
		log.Printf("Deployment %s deleted successfully", resourceName)
	}
	return nil
}

// deleteAllManagedServicePVCs deletes all PVCs for managed services (always StatefulSet PVCs)
func deleteAllManagedServicePVCs(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	if !RequiresPersistentStorage(service.ManagedType) {
		log.Printf("Service %s does not require persistent storage, skipping PVC deletion", service.ManagedType)
		return nil
	}
	
	// All managed services use StatefulSet, so always delete StatefulSet PVCs
	return deleteStatefulSetPVCs(ctx, k8sClient, service)
}

// deleteStatefulSetPVCs deletes PVCs created by StatefulSet volume claim templates
func deleteStatefulSetPVCs(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	
	// StatefulSet creates PVCs with pattern: <volumeClaimTemplate>-<statefulset-name>-<ordinal>
	// Our volume claim template is named "data", so PVC will be "data-<resourceName>-0"
	pvcName := fmt.Sprintf("data-%s-0", resourceName)
	
	err := k8sClient.Clientset.CoreV1().PersistentVolumeClaims(service.EnvironmentID).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete StatefulSet PVC %s: %v", pvcName, err)
	}
	
	if err == nil {
		log.Printf("StatefulSet PVC %s deleted successfully", pvcName)
	} else {
		log.Printf("StatefulSet PVC %s not found or already deleted", pvcName)
	}
	
	// Also check for any manual PVC that might exist (legacy cleanup)
	legacyPVCName := fmt.Sprintf("%s-data", resourceName)
	err = k8sClient.Clientset.CoreV1().PersistentVolumeClaims(service.EnvironmentID).Delete(ctx, legacyPVCName, metav1.DeleteOptions{})
	if err == nil {
		log.Printf("Legacy PVC %s deleted successfully", legacyPVCName)
	}
	
	return nil
}

// deleteDeploymentPVC deletes standalone PVC for Deployment-based managed services
func deleteDeploymentPVC(ctx context.Context, k8sClient *kubernetes.Client, service models.Service) error {
	resourceName := GetResourceName(service)
	pvcName := fmt.Sprintf("%s-data", resourceName)
	
	err := k8sClient.Clientset.CoreV1().PersistentVolumeClaims(service.EnvironmentID).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		log.Printf("Warning: Failed to delete Deployment PVC %s: %v", pvcName, err)
		return nil // Don't fail for missing standalone PVCs
	}
	
	if err == nil {
		log.Printf("Deployment PVC %s deleted successfully", pvcName)
	} else {
		log.Printf("Deployment PVC %s not found or already deleted", pvcName)
	}
	
	return nil
}

// DeleteAllResourcesInNamespace deletes all resources in a namespace (for environment cleanup)
func DeleteAllResourcesInNamespace(environmentID string) error {
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	ctx := context.Background()
	
	log.Printf("Starting cleanup of all resources in namespace: %s", environmentID)
	
	// Delete all resources in order
	var deletionErrors []string
	
	// Delete HPAs
	if err := deleteAllHPAsInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("HPAs: %v", err))
	}
	
	// Delete all TCP IngressRoutes
	if err := deleteAllTCPIngressRoutesInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("TCP IngressRoutes: %v", err))
	}
	
	// Delete all Ingresses
	if err := deleteAllIngressesInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("Ingresses: %v", err))
	}
	
	// Delete all Services (except default kubernetes service)
	if err := deleteAllServicesInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("Services: %v", err))
	}
	
	// Delete all StatefulSets
	if err := deleteAllStatefulSetsInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("StatefulSets: %v", err))
	}
	
	// Delete all Deployments
	if err := deleteAllDeploymentsInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("Deployments: %v", err))
	}
	
	// Delete all PVCs
	if err := deleteAllPVCsInNamespace(ctx, k8sClient, environmentID); err != nil {
		deletionErrors = append(deletionErrors, fmt.Sprintf("PVCs: %v", err))
	}
	
	if len(deletionErrors) > 0 {
		return fmt.Errorf("some resources failed to delete: %v", deletionErrors)
	}
	
	log.Printf("Successfully deleted all resources in namespace: %s", environmentID)
	return nil
}

// Helper functions for namespace-wide cleanup
func deleteAllHPAsInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	return k8sClient.Clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

func deleteAllTCPIngressRoutesInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	gvr := schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutetcps",
	}

	dynamicClient := k8sClient.DynamicClient.Resource(gvr).Namespace(namespace)
	return dynamicClient.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

func deleteAllIngressesInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	return k8sClient.Clientset.NetworkingV1().Ingresses(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

func deleteAllServicesInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	// List services first to avoid deleting kubernetes default service
	services, err := k8sClient.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, svc := range services.Items {
		// Skip default kubernetes service
		if svc.Name == "kubernetes" {
			continue
		}
		
		err := k8sClient.Clientset.CoreV1().Services(namespace).Delete(ctx, svc.Name, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			log.Printf("Warning: Failed to delete Service %s: %v", svc.Name, err)
		}
	}
	
	return nil
}

func deleteAllStatefulSetsInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	return k8sClient.Clientset.AppsV1().StatefulSets(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

func deleteAllDeploymentsInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	return k8sClient.Clientset.AppsV1().Deployments(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

func deleteAllPVCsInNamespace(ctx context.Context, k8sClient *kubernetes.Client, namespace string) error {
	return k8sClient.Clientset.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
}

// CleanupOrphanedResources finds and deletes resources that don't belong to any active service
func CleanupOrphanedResources(environmentID string, activeServices []models.Service) error {
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	ctx := context.Background()
	
	// Build list of expected resource names
	expectedResourceNames := make(map[string]bool)
	for _, service := range activeServices {
		resourceName := GetResourceName(service)
		expectedResourceNames[resourceName] = true
		
		// For managed services, also track secondary resources
		if service.Type == models.ServiceTypeManaged {
			exposureConfigs := GetManagedServiceExposureConfig(service.ManagedType)
			for _, config := range exposureConfigs {
				if config.Name != "primary" {
					secondaryName := fmt.Sprintf("%s-%s", resourceName, config.Name)
					expectedResourceNames[secondaryName] = true
				}
			}
		}
	}
	
	log.Printf("Starting orphaned resource cleanup in namespace: %s", environmentID)
	log.Printf("Expected resources: %v", expectedResourceNames)
	
	// Find and delete orphaned resources
	if err := cleanupOrphanedTCPIngressRoutes(ctx, k8sClient, environmentID, expectedResourceNames); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned TCP IngressRoutes: %v", err)
	}
	
	if err := cleanupOrphanedIngresses(ctx, k8sClient, environmentID, expectedResourceNames); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned Ingresses: %v", err)
	}
	
	if err := cleanupOrphanedServices(ctx, k8sClient, environmentID, expectedResourceNames); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned Services: %v", err)
	}
	
	if err := cleanupOrphanedWorkloads(ctx, k8sClient, environmentID, expectedResourceNames); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned Workloads: %v", err)
	}
	
	if err := cleanupOrphanedPVCs(ctx, k8sClient, environmentID, expectedResourceNames); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned PVCs: %v", err)
	}
	
	log.Printf("Orphaned resource cleanup completed for namespace: %s", environmentID)
	return nil
}

func cleanupOrphanedTCPIngressRoutes(ctx context.Context, k8sClient *kubernetes.Client, namespace string, expected map[string]bool) error {
	gvr := schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutetcps",
	}

	dynamicClient := k8sClient.DynamicClient.Resource(gvr).Namespace(namespace)
	
	tcpRoutes, err := dynamicClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, route := range tcpRoutes.Items {
		if !expected[route.GetName()] {
			log.Printf("Deleting orphaned TCP IngressRoute: %s", route.GetName())
			err := dynamicClient.Delete(ctx, route.GetName(), metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete orphaned TCP IngressRoute %s: %v", route.GetName(), err)
			}
		}
	}
	
	return nil
}

func cleanupOrphanedIngresses(ctx context.Context, k8sClient *kubernetes.Client, namespace string, expected map[string]bool) error {
	ingresses, err := k8sClient.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, ingress := range ingresses.Items {
		if !expected[ingress.Name] {
			log.Printf("Deleting orphaned Ingress: %s", ingress.Name)
			err := k8sClient.Clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, ingress.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete orphaned Ingress %s: %v", ingress.Name, err)
			}
		}
	}
	
	return nil
}

func cleanupOrphanedServices(ctx context.Context, k8sClient *kubernetes.Client, namespace string, expected map[string]bool) error {
	services, err := k8sClient.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, service := range services.Items {
		// Skip default kubernetes service
		if service.Name == "kubernetes" {
			continue
		}
		
		if !expected[service.Name] {
			log.Printf("Deleting orphaned Service: %s", service.Name)
			err := k8sClient.Clientset.CoreV1().Services(namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete orphaned Service %s: %v", service.Name, err)
			}
		}
	}
	
	return nil
}

func cleanupOrphanedWorkloads(ctx context.Context, k8sClient *kubernetes.Client, namespace string, expected map[string]bool) error {
	// Check StatefulSets
	statefulSets, err := k8sClient.Clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, sts := range statefulSets.Items {
		if !expected[sts.Name] {
			log.Printf("Deleting orphaned StatefulSet: %s", sts.Name)
			err := k8sClient.Clientset.AppsV1().StatefulSets(namespace).Delete(ctx, sts.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete orphaned StatefulSet %s: %v", sts.Name, err)
			}
		}
	}
	
	// Check Deployments
	deployments, err := k8sClient.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, deploy := range deployments.Items {
		if !expected[deploy.Name] {
			log.Printf("Deleting orphaned Deployment: %s", deploy.Name)
			err := k8sClient.Clientset.AppsV1().Deployments(namespace).Delete(ctx, deploy.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete orphaned Deployment %s: %v", deploy.Name, err)
			}
		}
	}
	
	return nil
}

func cleanupOrphanedPVCs(ctx context.Context, k8sClient *kubernetes.Client, namespace string, expected map[string]bool) error {
	pvcs, err := k8sClient.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	
	for _, pvc := range pvcs.Items {
		// Check if PVC belongs to any expected resource
		isOrphaned := true
		
		// Check against expected resource names and their PVC patterns
		for expectedName := range expected {
			// StatefulSet PVC pattern: data-<resourceName>-0
			if pvc.Name == fmt.Sprintf("data-%s-0", expectedName) {
				isOrphaned = false
				break
			}
			// Deployment PVC pattern: <resourceName>-data
			if pvc.Name == fmt.Sprintf("%s-data", expectedName) {
				isOrphaned = false
				break
			}
		}
		
		if isOrphaned {
			log.Printf("Deleting orphaned PVC: %s", pvc.Name)
			err := k8sClient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Printf("Warning: Failed to delete orphaned PVC %s: %v", pvc.Name, err)
			}
		}
	}
	
	return nil
}