package utils

import (
	"context"
	"fmt"

	"github.com/pendeploy-simple/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	// RegistryNamespace is the namespace for registry deployments
	RegistryNamespace = "registry"
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

// DeployRegistry deploys a registry to Kubernetes and returns the pod name when available
func (d *RegistryDeployer) DeployRegistry(ctx context.Context, registry models.Registry) (string, error) {
	// Ensure namespace exists
	if err := d.ensureNamespaceExists(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure namespace exists: %w", err)
	}

	// Create ConfigMap
	if err := d.createConfigMap(ctx, registry); err != nil {
		return "", fmt.Errorf("failed to create config map: %w", err)
	}

	// Create Secret
	if err := d.createSecret(ctx, registry); err != nil {
		return "", fmt.Errorf("failed to create secret: %w", err)
	}
	
	// Create PVC first
	if err := d.createPVC(ctx, registry); err != nil {
		return "", fmt.Errorf("failed to create persistent volume claim: %w", err)
	}

	// Create Deployment
	if err := d.createDeployment(ctx, registry); err != nil {
		return "", fmt.Errorf("failed to create deployment: %w", err)
	}

	// Create Service
	if err := d.createService(ctx, registry); err != nil {
		return "", fmt.Errorf("failed to create service: %w", err)
	}

	// Wait for pod to be created and get its name
	return d.waitForRegistryPod(ctx, registry)
}

// UpdateRegistry updates a registry in Kubernetes
func (d *RegistryDeployer) UpdateRegistry(ctx context.Context, registry models.Registry) error {
	// Update ConfigMap
	if err := d.updateConfigMap(ctx, registry); err != nil {
		return fmt.Errorf("failed to update config map: %w", err)
	}

	// Update Secret if there's a password change
	if registry.Password != "" {
		if err := d.updateSecret(ctx, registry); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}
	}

	// Update Deployment (will trigger a rolling update)
	if err := d.updateDeployment(ctx, registry); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	// Service doesn't typically need updates since it's just mapping ports

	return nil
}

// DeleteRegistry deletes a registry from Kubernetes
func (d *RegistryDeployer) DeleteRegistry(ctx context.Context, registryID string) error {
	resourceName := d.getRegistryResourceName(registryID)
	var errs []string
	
	// Get list of pods with this registry ID to ensure we delete the correct resources
	podList, err := d.clientset.CoreV1().Pods(RegistryNamespace).List(ctx, metav1.ListOptions{
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
		deleteErr := d.clientset.CoreV1().Pods(RegistryNamespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if deleteErr != nil {
			fmt.Printf("Error deleting pod %s: %v\n", pod.Name, deleteErr)
		} else {
			fmt.Printf("Successfully deleted pod %s\n", pod.Name)
		}
	}

	// Delete service
	if err := d.clientset.CoreV1().Services(RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		// Log error and collect it
		errs = append(errs, fmt.Sprintf("Error deleting service for registry %s: %v", registryID, err))
		fmt.Printf("Error deleting service for registry %s: %v\n", registryID, err)
	} else {
		fmt.Printf("Successfully deleted service %s\n", resourceName)
	}

	// Delete deployment
	if err := d.clientset.AppsV1().Deployments(RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		errs = append(errs, fmt.Sprintf("Error deleting deployment for registry %s: %v", registryID, err))
		fmt.Printf("Error deleting deployment for registry %s: %v\n", registryID, err)
	} else {
		fmt.Printf("Successfully deleted deployment %s\n", resourceName)
	}

	// Delete secret
	if err := d.clientset.CoreV1().Secrets(RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		errs = append(errs, fmt.Sprintf("Error deleting secret for registry %s: %v", registryID, err))
		fmt.Printf("Error deleting secret for registry %s: %v\n", registryID, err)
	} else {
		fmt.Printf("Successfully deleted secret %s\n", resourceName)
	}

	// Delete configmap
	if err := d.clientset.CoreV1().ConfigMaps(RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		errs = append(errs, fmt.Sprintf("Error deleting configmap for registry %s: %v", registryID, err))
		fmt.Printf("Error deleting configmap for registry %s: %v\n", registryID, err)
	} else {
		fmt.Printf("Successfully deleted configmap %s\n", resourceName)
	}
	
	// Delete PVC
	if err := d.clientset.CoreV1().PersistentVolumeClaims(RegistryNamespace).Delete(ctx, resourceName, metav1.DeleteOptions{}); err != nil {
		errs = append(errs, fmt.Sprintf("Error deleting PVC for registry %s: %v", registryID, err))
		fmt.Printf("Error deleting PVC for registry %s: %v\n", registryID, err)
	} else {
		fmt.Printf("Successfully deleted PVC %s\n", resourceName)
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("Failed to delete some Kubernetes resources: %v", errs)
	}
	return nil
}

// Helper methods

func (d *RegistryDeployer) ensureNamespaceExists(ctx context.Context) error {
	_, err := d.clientset.CoreV1().Namespaces().Get(ctx, RegistryNamespace, metav1.GetOptions{})
	if err == nil {
		// Namespace already exists
		return nil
	}

	// Create namespace if it doesn't exist
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: RegistryNamespace,
		},
	}
	_, err = d.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	return err
}

func (d *RegistryDeployer) createConfigMap(ctx context.Context, registry models.Registry) error {
	configMapData := map[string]string{
		"registry.yml": fmt.Sprintf(`
version: 0.1
log:
  level: info
storage:
  filesystem:
    rootdirectory: /var/lib/registry
  delete:
    enabled: true
http:
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
`),
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.getRegistryResourceName(registry.ID),
			Namespace: RegistryNamespace,
			Labels: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
		},
		Data: configMapData,
	}

	_, err := d.clientset.CoreV1().ConfigMaps(RegistryNamespace).Create(ctx, configMap, metav1.CreateOptions{})
	return err
}

func (d *RegistryDeployer) updateConfigMap(ctx context.Context, registry models.Registry) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get current config map
		configMap, err := d.clientset.CoreV1().ConfigMaps(RegistryNamespace).Get(ctx, d.getRegistryResourceName(registry.ID), metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Update config map data
		configMap.Data = map[string]string{
			"registry.yml": fmt.Sprintf(`
version: 0.1
log:
  level: info
storage:
  filesystem:
    rootdirectory: /var/lib/registry
  delete:
    enabled: true
http:
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
`),
		}

		_, err = d.clientset.CoreV1().ConfigMaps(RegistryNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
		return err
	})
}

func (d *RegistryDeployer) createSecret(ctx context.Context, registry models.Registry) error {
	// Tetapkan username secara konsisten
	username := "admin"
	
	// Generate htpasswd format untuk autentikasi Docker Registry
	// Format: username:password
	htpasswd := fmt.Sprintf("%s:%s", username, registry.Password)

	// Generate internal Kubernetes URL for the registry service
	// Format: registry-service-name.namespace.svc.cluster.local
	// This is the standard Kubernetes DNS format for accessing services within the cluster
	resourceName := d.getRegistryResourceName(registry.ID)
	registryURL := fmt.Sprintf("%s.%s.svc.cluster.local:5000", resourceName, RegistryNamespace)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: RegistryNamespace,
			Labels: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"htpasswd":  htpasswd,  // Dibutuhkan untuk autentikasi Docker Registry
			"url":      registryURL,
			"username": username,
			"password": registry.Password, // Password mentah untuk konsistensi dengan database
		},
	}

	_, err := d.clientset.CoreV1().Secrets(RegistryNamespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (d *RegistryDeployer) updateSecret(ctx context.Context, registry models.Registry) error {
	resourceName := d.getRegistryResourceName(registry.ID)

	// Get existing secret
	secret, err := d.clientset.CoreV1().Secrets(RegistryNamespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Update data
	if secret.StringData == nil {
		secret.StringData = map[string]string{}
	}

	// Only update password if provided
	if registry.Password != "" {
		username := "admin"
		secret.StringData["password"] = registry.Password
		secret.StringData["htpasswd"] = fmt.Sprintf("%s:%s", username, registry.Password)
	}

	// URL is managed by Kubernetes and should not be updated from external inputs
	// We'll preserve the original URL that was generated

	_, err = d.clientset.CoreV1().Secrets(RegistryNamespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

// createPVC creates a persistent volume claim for registry data with 10Gi storage
func (d *RegistryDeployer) createPVC(ctx context.Context, registry models.Registry) error {
	// Log the PVC creation
	fmt.Printf("Creating PVC with name %s in namespace %s\n", d.getRegistryResourceName(registry.ID), RegistryNamespace)
	
	// Create the PVC directly with the standard structure
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.getRegistryResourceName(registry.ID),
			Namespace: RegistryNamespace,
			Labels: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
		},
	}
	
	// Add access mode
	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	pvc.Spec.AccessModes = accessModes
	
	// Add storage request - 10Gi
	pvc.Spec.Resources.Requests = make(corev1.ResourceList)
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("5Gi")
	
	_, err := d.clientset.CoreV1().PersistentVolumeClaims(RegistryNamespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

func (d *RegistryDeployer) createDeployment(ctx context.Context, registry models.Registry) error {
	var replicas int32 = 1
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.getRegistryResourceName(registry.ID),
			Namespace: RegistryNamespace,
			Labels: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":        "registry",
					"registry-id": registry.ID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "registry",
						"registry-id": registry.ID,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "registry",
							Image: "registry:2",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 5000,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/docker/registry",
									ReadOnly:  true,
								},
								{
									Name:      "data",
									MountPath: "/var/lib/registry",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intToQuantity(5000),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:      30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intToQuantity(5000),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:      10,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: d.getRegistryResourceName(registry.ID),
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "registry.yml",
											Path: "config.yml",
										},
									},
								},
							},
						},
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: d.getRegistryResourceName(registry.ID),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := d.clientset.AppsV1().Deployments(RegistryNamespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func (d *RegistryDeployer) updateDeployment(ctx context.Context, registry models.Registry) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get current deployment
		deployment, err := d.clientset.AppsV1().Deployments(RegistryNamespace).Get(ctx, d.getRegistryResourceName(registry.ID), metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Update any deployment spec fields that need to be changed
		// This could include resource limits, etc.
		// For now, let's just update the annotations to trigger a rolling update
		
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		deployment.Spec.Template.Annotations["pendeploy.com/update-timestamp"] = fmt.Sprintf("%d", metav1.Now().Unix())

		_, err = d.clientset.AppsV1().Deployments(RegistryNamespace).Update(ctx, deployment, metav1.UpdateOptions{})
		return err
	})
}

func (d *RegistryDeployer) createService(ctx context.Context, registry models.Registry) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.getRegistryResourceName(registry.ID),
			Namespace: RegistryNamespace,
			Labels: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       5000,
					TargetPort: intToQuantity(5000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err := d.clientset.CoreV1().Services(RegistryNamespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// waitForRegistryPod waits for a pod to be created and returns its name
func (d *RegistryDeployer) waitForRegistryPod(ctx context.Context, registry models.Registry) (string, error) {
	labelSelector := fmt.Sprintf("app=registry,registry-id=%s", registry.ID)
	
	// Poll until a pod is found or timeout
	var podName string
	err := retry.OnError(retry.DefaultRetry, func(error) bool { return true }, func() error {
		pods, err := d.clientset.CoreV1().Pods(RegistryNamespace).List(ctx, metav1.ListOptions{
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

// getRegistryResourceName returns a consistent name for Kubernetes resources
func (d *RegistryDeployer) getRegistryResourceName(registryID string) string {
	// Use the full ID as requested to preserve ID for tracking
	return fmt.Sprintf("registry-%s", registryID)
}

// GetRegistrySecret retrieves the secret for a given registry from Kubernetes
func (d *RegistryDeployer) GetRegistrySecret(ctx context.Context, registryID string) (*corev1.Secret, error) {
	resourceName := d.getRegistryResourceName(registryID)
	
	// Get secret from Kubernetes
	secret, err := d.clientset.CoreV1().Secrets(RegistryNamespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get registry secret: %w", err)
	}
	
	return secret, nil
}

// intToQuantity converts an int to an IntOrString for Kubernetes API
func intToQuantity(val int) intstr.IntOrString {
	return intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(val),
	}
}
