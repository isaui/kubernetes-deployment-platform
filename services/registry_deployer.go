package services

import (
	"context"
	"fmt"

	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RegistryDeployer handles Kubernetes operations for deploying registries
type RegistryDeployer struct {
	clientset *kubernetes.Clientset
}

// NewRegistryDeployer creates a new registry deployer instance
func NewRegistryDeployer(clientset *kubernetes.Clientset) *RegistryDeployer {
	return &RegistryDeployer{
		clientset: clientset,
	}
}

// getRegistryURL returns the public hostname for this registry.
func (d *RegistryDeployer) getRegistryURL(registryID string) string {
	return utils.GetRegistryHostname(registryID)
}

// DeployRegistry deploys a registry to Kubernetes and returns the pod name and URL when available
func (d *RegistryDeployer) DeployRegistry(ctx context.Context, registry models.Registry) (string, string, error) {
	// Ensure namespace exists
	if err := utils.EnsureNamespaceExists(utils.RegistryNamespace); err != nil {
		return "", "", fmt.Errorf("failed to ensure namespace exists: %w", err)
	}
	// Create PVC first
	if err := utils.CreatePVC(ctx, registry, utils.RegistryNamespace, d.clientset); err != nil {
		return "", "", fmt.Errorf("failed to create persistent volume claim: %w", err)
	}

	// Create Service
	if err := utils.CreateRegistryService(ctx, utils.RegistryNamespace, registry, d.clientset); err != nil {
		return "", "", fmt.Errorf("failed to create service: %w", err)
	}

	// Create Deployment
	if err := utils.CreateRegistryDeployment(ctx, utils.RegistryNamespace, registry, d.clientset); err != nil {
		return "", "", fmt.Errorf("failed to create deployment: %w", err)
	}

	if err := utils.CreateRegistryIngress(ctx, utils.RegistryNamespace, registry, d.clientset); err != nil {
		return "", "", fmt.Errorf("failed to create ingress: %w", err)
	}

	// Wait for pod to be created and get its name
	podName, err := utils.WaitForRegistryPod(ctx, registry, utils.RegistryNamespace, d.clientset)
	if err != nil {
		return "", "", err
	}

	registryURL := d.getRegistryURL(registry.ID)

	// Log the URL for debugging
	fmt.Printf("Setting registry URL to public hostname: %s\n", registryURL)

	// Return URL without scheme as requested
	return podName, registryURL, nil
}

// UpdateRegistry updates a registry in Kubernetes
func (d *RegistryDeployer) UpdateRegistry(ctx context.Context, registry models.Registry) error {
	// Update Deployment (will trigger a rolling update)
	if err := utils.UpdateDeployment(ctx, registry, d.clientset, utils.RegistryNamespace); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	if err := utils.CreateRegistryService(ctx, utils.RegistryNamespace, registry, d.clientset); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	if err := utils.CreateRegistryIngress(ctx, utils.RegistryNamespace, registry, d.clientset); err != nil {
		return fmt.Errorf("failed to update ingress: %w", err)
	}

	return nil
}

// DeleteRegistry deletes a registry from Kubernetes
func (d *RegistryDeployer) DeleteRegistry(ctx context.Context, registryID string) error {
	resourceName := utils.GetRegistryResourceName(registryID)
	var errs []string

	// Get list of pods with this registry ID to ensure we delete the correct resources
	podList, err := d.clientset.CoreV1().Pods(utils.RegistryNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("registry-id=%s", registryID),
	})
	if err != nil {
		errs = append(errs, fmt.Sprintf("Error listing pods for registry %s: %v", registryID, err))
	}

	// Log pod names for debugging
	fmt.Printf("Found %d pods for registry-id=%s\n", len(podList.Items), registryID)
	for _, pod := range podList.Items {
		fmt.Printf("Pod found: %s\n", pod.Name)

		// Delete any found pods explicitly
		deleteErr := d.clientset.CoreV1().Pods(utils.RegistryNamespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if deleteErr != nil {
			fmt.Printf("Error deleting pod %s: %v\n", pod.Name, deleteErr)
		} else {
			fmt.Printf("Successfully deleted pod %s\n", pod.Name)
		}
	}

	// Delete service
	if err := d.clientset.CoreV1().Services(utils.RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Sprintf("Error deleting service for registry %s: %v", registryID, err))
			fmt.Printf("Error deleting service for registry %s: %v\n", registryID, err)
		}
	} else {
		fmt.Printf("Successfully deleted service %s\n", resourceName)
	}

	// Delete deployment
	if err := d.clientset.AppsV1().Deployments(utils.RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Sprintf("Error deleting deployment for registry %s: %v", registryID, err))
			fmt.Printf("Error deleting deployment for registry %s: %v\n", registryID, err)
		}
	} else {
		fmt.Printf("Successfully deleted deployment %s\n", resourceName)
	}

	if err := d.clientset.NetworkingV1().Ingresses(utils.RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Sprintf("Error deleting ingress for registry %s: %v", registryID, err))
			fmt.Printf("Error deleting ingress for registry %s: %v\n", registryID, err)
		}
	} else {
		fmt.Printf("Successfully deleted ingress %s\n", resourceName)
	}

	// Delete PVC
	if err := d.clientset.CoreV1().PersistentVolumeClaims(utils.RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			errs = append(errs, fmt.Sprintf("Error deleting PVC for registry %s: %v", registryID, err))
			fmt.Printf("Error deleting PVC for registry %s: %v\n", registryID, err)
		}
	} else {
		fmt.Printf("Successfully deleted PVC %s\n", resourceName)
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("Failed to delete some Kubernetes resources: %v", errs)
	}
	return nil
}
