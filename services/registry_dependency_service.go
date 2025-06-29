package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/utils"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
const (
	CAConfigMapName = "pendeploy-registry-ca"
)
// RegistryDependencyService handles setup and management of required images for a registry
type RegistryDependencyService struct {
	kubeClient *kubernetes.Client
}

// DependencyImage represents a required image for the build system
type DependencyImage struct {
	Name         string
	SourceImage  string
	TargetTag    string
	Description  string
}

// NewRegistryDependencyService creates a new registry dependency service
func NewRegistryDependencyService() *RegistryDependencyService {
	client, err := kubernetes.NewClient()
	if err != nil {
		log.Printf("Warning: Could not create Kubernetes client for dependency service: %v", err)
	}
	
	return &RegistryDependencyService{
		kubeClient: client,
	}
}

// GetRequiredImages returns the list of images required for the build system
func (s *RegistryDependencyService) GetRequiredImages() []DependencyImage {
	return []DependencyImage{
		{
			Name:        "alpine-git",
			SourceImage: "alpine/git:2.43.0",
			TargetTag:   "alpine-git:2.43.0",
			Description: "Lightweight Git client for source code cloning",
		},
		{
			Name:        "kaniko-executor",
			SourceImage: "gcr.io/kaniko-project/executor:v1.23.2",
			TargetTag:   "kaniko-executor:v1.23.2",
			Description: "Daemonless container image builder",
		},
		{
			Name:        "nixpacks-ready",
			SourceImage: "nixos/nix:2.18.1",
			TargetTag:   "nixpacks-ready:v1.0.0",
			Description: "Custom image with nixpacks CLI pre-installed",
		},
	}
}

// SetupRegistryDependencies ensures all required images are available in the target registry
func (s *RegistryDependencyService) SetupRegistryDependencies(ctx context.Context, registry models.Registry) error {
	if s.kubeClient == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	log.Printf("Setting up dependencies for registry %s (%s)", registry.Name, registry.URL)

	// Wait for registry to be ready first
	log.Printf("Waiting for registry to be ready...")
	err := s.waitForRegistryReady(ctx, registry, 3*time.Minute)
	if err != nil {
		return fmt.Errorf("registry not ready: %v", err)
	}
	log.Printf("Registry is ready, proceeding with dependency setup")

	// Ensure namespace exists
	err = s.ensureNamespaceExists(ctx, "build-and-deploy")
	if err != nil {
		return fmt.Errorf("failed to ensure namespace exists: %v", err)
	}

	// Get required images
	images := s.GetRequiredImages()
	
	// Process each image - ALL USING KANIKO NOW!
	for _, img := range images {
		log.Printf("Building dependency image with Kaniko: %s", img.Name)
		
		err := s.buildImageWithKaniko(ctx, registry, img)
		if err != nil {
			log.Printf("Failed to build image %s with Kaniko: %v", img.Name, err)
			return fmt.Errorf("failed to build image %s: %v", img.Name, err)
		}
	}

	log.Printf("Successfully set up all dependencies for registry %s", registry.Name)
	return nil
}

// waitForRegistryReady waits for registry to be accessible
func (s *RegistryDependencyService) waitForRegistryReady(ctx context.Context, registry models.Registry, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for registry to be ready")
		case <-ticker.C:
			// Test registry connectivity using a simple curl pod
			testPodName := fmt.Sprintf("registry-test-%s", utils.GenerateShortID())
			
			testPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPodName,
					Namespace: "build-and-deploy",
					Labels: map[string]string{
						"app": "pendeploy-test",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "curlimages/curl:latest",
							Command: []string{
								"sh", "-c",
								fmt.Sprintf("curl -f -k %s/v2/ && echo 'Registry ready'", registry.URL),
							},
						},
					},
				},
			}
			
			// Create test pod
			_, err := s.kubeClient.Clientset.CoreV1().Pods("build-and-deploy").Create(ctx, testPod, metav1.CreateOptions{})
			if err != nil {
				log.Printf("Failed to create test pod: %v", err)
				continue
			}
			
			// Wait for pod completion
			time.Sleep(6 * time.Second)
			
			// Check if test passed
			pod, err := s.kubeClient.Clientset.CoreV1().Pods("build-and-deploy").Get(ctx, testPodName, metav1.GetOptions{})
			if err == nil && pod.Status.Phase == corev1.PodSucceeded {
				// Cleanup test pod
				s.kubeClient.Clientset.CoreV1().Pods("build-and-deploy").Delete(ctx, testPodName, metav1.DeleteOptions{})
				log.Printf("Registry connectivity test passed")
				return nil
			}
			
			// Cleanup failed test pod
			s.kubeClient.Clientset.CoreV1().Pods("build-and-deploy").Delete(ctx, testPodName, metav1.DeleteOptions{})
			log.Printf("Registry not ready yet, retrying...")
		}
	}
}

// buildImageWithKaniko builds and pushes ALL images using Kaniko
func (s *RegistryDependencyService) buildImageWithKaniko(ctx context.Context, registry models.Registry, img DependencyImage) error {
	log.Printf("Building image with Kaniko: %s -> %s/%s", img.SourceImage, registry.URL, img.TargetTag)

	// Generate unique job name
	jobName := fmt.Sprintf("build-%s-%s", img.Name, utils.GenerateShortID())
	
	// Create build job
	job, err := s.createRegistryDependencyKanikoJob(registry, img, jobName)
	if err != nil {
		return fmt.Errorf("failed to create Kaniko job: %v", err)
	}

	// Submit job to Kubernetes
	_, err = s.kubeClient.Clientset.BatchV1().Jobs("build-and-deploy").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to submit Kaniko job: %v", err)
	}

	log.Printf("Kaniko build job %s submitted successfully", jobName)
	return nil
}

// createRegistryDependencyKanikoJob creates a Kubernetes job for building ALL images with Kaniko
func (s *RegistryDependencyService) createRegistryDependencyKanikoJob(registry models.Registry, img DependencyImage, jobName string) (*batchv1.Job, error) {
	// Generate dockerfile content based on image type
	var dockerfileContent string
	
	switch img.Name {
	case "alpine-git":
		// Simple retagging - just use the base image
		dockerfileContent = fmt.Sprintf(`FROM %s
# Alpine Git image ready for use
WORKDIR /workspace`, img.SourceImage)
		
	case "kaniko-executor":
		// Simple retagging - just use the base image
		dockerfileContent = fmt.Sprintf(`FROM %s
# Kaniko executor ready for use
WORKDIR /workspace`, img.SourceImage)
		
	case "nixpacks-ready":
		// Custom build with nixpacks
		dockerfileContent = fmt.Sprintf(`FROM %s
RUN nix-env -iA nixpkgs.nixpacks
RUN nix-collect-garbage -d
RUN nixpacks --version
WORKDIR /workspace`, img.SourceImage)
		
	default:
		return nil, fmt.Errorf("unknown image: %s", img.Name)
	}

	// Clean registry URL for kaniko
	cleanRegistryURL := utils.CleanRegistryURL(registry.URL)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "build-and-deploy",
			Labels: map[string]string{
				"app":          "pendeploy",
				"job-type":     "kaniko-build",
				"registry-id":  registry.ID,
				"image-name":   img.Name,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(0), // No retries for cleaner logs
			TTLSecondsAfterFinished: int32Ptr(600), // 10 minutes cleanup
			ActiveDeadlineSeconds:   int64Ptr(1800), // 30 minutes timeout
			
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":          "pendeploy",
						"job-type":     "kaniko-build",
						"registry-id":  registry.ID,
						"image-name":   img.Name,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					
					InitContainers: []corev1.Container{
						{
							Name:  "dockerfile-generator",
							Image: "alpine:3.19.1",
							Command: []string{
								"sh", "-c",
								fmt.Sprintf("echo '%s' > /workspace/Dockerfile && echo 'Dockerfile generated for %s' && cat /workspace/Dockerfile", 
									dockerfileContent, img.Name),
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
					
					Containers: []corev1.Container{
						{
							Name:  "kaniko-builder",
							Image: "gcr.io/kaniko-project/executor:v1.23.2",
							Args: []string{
								"--context=/workspace",
								"--dockerfile=/workspace/Dockerfile",
								fmt.Sprintf("--destination=%s/%s", cleanRegistryURL, img.TargetTag),
								"--cache=false",
								"--cleanup",
								"--verbosity=info",
								"--insecure",
								"--force",
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("360Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"), // Increase for faster builds
									corev1.ResourceMemory: resource.MustParse("2Gi"),   // More memory for large images
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/kaniko/.docker/config.json",
								},
								{
									Name:  "KANIKO_DIR",
									Value: "/kaniko",
								},
							},
						},
					},
					
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}, nil
}
// ValidateDependencies checks if all required images are available in the registry
func (s *RegistryDependencyService) ValidateDependencies(ctx context.Context, registry models.Registry) ([]string, error) {
	var missingImages []string
	
	// Get required images
	images := s.GetRequiredImages()
	
	// Check each image (this would typically involve registry API calls)
	for _, img := range images {
		// For now, we'll assume they need to be checked via registry API
		// TODO: Implement actual registry API validation
		log.Printf("Would validate image: %s/%s", registry.URL, img.TargetTag)
	}
	
	return missingImages, nil
}

// CleanupDependencyJobs removes old dependency setup jobs
func (s *RegistryDependencyService) CleanupDependencyJobs(ctx context.Context, registryID string, olderThan time.Duration) error {
	if s.kubeClient == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	// List jobs related to this registry
	jobs, err := s.kubeClient.Clientset.BatchV1().Jobs("build-and-deploy").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("registry-id=%s", registryID),
	})
	if err != nil {
		return fmt.Errorf("failed to list dependency jobs: %v", err)
	}

	cutoffTime := time.Now().Add(-olderThan)
	
	for _, job := range jobs.Items {
		if job.CreationTimestamp.Before(&metav1.Time{Time: cutoffTime}) {
			log.Printf("Cleaning up old dependency job: %s", job.Name)
			
			err := s.kubeClient.Clientset.BatchV1().Jobs("build-and-deploy").Delete(
				ctx, job.Name, metav1.DeleteOptions{},
			)
			if err != nil && !errors.IsNotFound(err) {
				log.Printf("Failed to delete job %s: %v", job.Name, err)
			}
		}
	}

	return nil
}



// ensureNamespaceExists creates namespace if it doesn't exist
func (s *RegistryDependencyService) ensureNamespaceExists(ctx context.Context, namespace string) error {
	_, err := s.kubeClient.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
					Labels: map[string]string{
						"app": "pendeploy",
					},
				},
			}
			_, err = s.kubeClient.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create namespace: %v", err)
			}
			log.Printf("Namespace %s created successfully", namespace)
		} else {
			return fmt.Errorf("failed to check namespace: %v", err)
		}
	}
	return nil
}
// Helper pointer functions
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }