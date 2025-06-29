package utils

import (
	"context"
	"fmt"

	"github.com/pendeploy-simple/models"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func CreateRegistryService(ctx context.Context, registryNamespace string, registry models.Registry, clientset *kubernetes.Clientset) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetRegistryResourceName(registry.ID),
			Namespace: registryNamespace,
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
					TargetPort: IntToQuantity(5000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err := clientset.CoreV1().Services(registryNamespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func CreateRegistryDeployment(ctx context.Context, registryNamespace string, registry models.Registry, clientset *kubernetes.Clientset) error {
	var replicas int32 = 1
	resourceName := GetRegistryResourceName(registry.ID)
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: registryNamespace,
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
							Env: []corev1.EnvVar{
								{
									Name:  "REGISTRY_STORAGE_DELETE_ENABLED",
									Value: "true",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
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
										Path:   "/v2/",
										Port:   IntToQuantity(5000),
										Scheme: corev1.URISchemeHTTP, // HTTP instead of HTTPS
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:      30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/v2/",
										Port:   IntToQuantity(5000),
										Scheme: corev1.URISchemeHTTP, // HTTP instead of HTTPS
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:      10,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: resourceName,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := clientset.AppsV1().Deployments(registryNamespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// CreatePVC creates a persistent volume claim for registry data with 5Gi storage
func CreatePVC(ctx context.Context, registry models.Registry, registryNamespace string, clientset *kubernetes.Clientset) error {
	// Log the PVC creation
	fmt.Printf("Creating PVC with name %s in namespace %s\n", GetRegistryResourceName(registry.ID), registryNamespace)
	
	// Create the PVC directly with the standard structure
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetRegistryResourceName(registry.ID),
			Namespace: registryNamespace,
			Labels: map[string]string{
				"app":        "registry",
				"registry-id": registry.ID,
			},
		},
	}
	
	// Add access mode
	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	pvc.Spec.AccessModes = accessModes
	
	// Add storage request - 5Gi
	pvc.Spec.Resources.Requests = make(corev1.ResourceList)
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse("5Gi")
	
	_, err := clientset.CoreV1().PersistentVolumeClaims(registryNamespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

func UpdateDeployment(ctx context.Context, registry models.Registry, clientset *kubernetes.Clientset, registryNamespace string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get current deployment
		deployment, err := clientset.AppsV1().Deployments(registryNamespace).Get(ctx, GetRegistryResourceName(registry.ID), metav1.GetOptions{})
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

		_, err = clientset.AppsV1().Deployments(registryNamespace).Update(ctx, deployment, metav1.UpdateOptions{})
		return err
	})
}