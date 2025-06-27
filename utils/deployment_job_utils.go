package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJobName(serviceID string, deploymentID string) string {
	return fmt.Sprintf("%s-%s-%s", serviceID, deploymentID, "build")
}

// GetJobNamespace returns the namespace for build jobs
func GetJobNamespace() string {
	return "build-and-deploy"
}

// BuildFromGit creates a Kubernetes job with init container for Git clone and main container for nixpacks build
// Returns the resulting image URL
func BuildFromGit(deployment models.Deployment, service models.Service, registry models.Registry) (string, error) {
    // Define the image tag - use "latest" if commitSHA is empty
    tagSuffix := "latest"
    if deployment.CommitSHA != "" {
        tagSuffix = deployment.CommitSHA
    }
    imageTag := fmt.Sprintf("%s/%s:%s", registry.URL, service.Name, tagSuffix)
    
    // Create Kubernetes client
    k8sClient, err := kubernetes.NewClient()
    if err != nil {
        return "", fmt.Errorf("failed to create Kubernetes client: %v", err)
    }
    
    // Create the job
    job, err := createGitBuildJob(deployment, service, registry, imageTag)
    if err != nil {
        return "", fmt.Errorf("failed to create job definition: %v", err)
    }
    
    // Submit the job to Kubernetes
    _, err = k8sClient.Clientset.BatchV1().Jobs(GetJobNamespace()).Create(
        context.Background(),
        job,
        metav1.CreateOptions{},
    )
    
    if err != nil {
        return "", fmt.Errorf("failed to create Kubernetes job: %v", err)
    }
    
    return imageTag, nil
}

// createGitBuildJob creates a job definition for building from Git
func createGitBuildJob(deployment models.Deployment, service models.Service, registry models.Registry, imageTag string) (*batchv1.Job, error) {
    // Create job name based on service and deployment ID
    jobName := GetJobName(service.ID, deployment.ID)
    
    // Parse branch (use default if empty)
    branch := service.Branch
    if branch == "" {
        branch = "main"
    }
    
    // Sanitize repository URL (ensure it ends with .git)
    repoURL := service.RepoURL
    if !strings.HasSuffix(repoURL, ".git") {
        repoURL = repoURL + ".git"
    }
    
    // Shared volume for git clone and build container
    sharedVolumeName := "source-code"
    
    // Generate Docker auth config for registry
    registryAuthConfig := generateDockerAuthConfig(registry)
    
    // Create job spec
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      jobName,
            Namespace: GetJobNamespace(),
            Labels: map[string]string{
                "app":         "pendeploy",
                "service-id":  service.ID,
                "deployment-id": deployment.ID,
            },
        },
        Spec: batchv1.JobSpec{
            // TTL to delete job after completion
            TTLSecondsAfterFinished: int32Ptr(3600), // 1 hour
            
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app":         "pendeploy",
                        "service-id":  service.ID,
                        "deployment-id": deployment.ID,
                    },
                },
                Spec: corev1.PodSpec{
                    // Don't restart on failure
                    RestartPolicy: corev1.RestartPolicyNever,
                    
                    // Init container for git clone
                    InitContainers: []corev1.Container{
                        {
                            Name:  "git-clone",
                            Image: "alpine/git:latest",
                            Command: []string{
                                "sh",
                                "-c",
                                fmt.Sprintf(
                                    "git clone --branch %s --single-branch %s /source && cd /source %s",
                                    branch,
                                    repoURL,
                                    getCheckoutCommand(deployment.CommitSHA),
                                ),
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      sharedVolumeName,
                                    MountPath: "/source",
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("300m"),
                                    corev1.ResourceMemory: resource.MustParse("256Mi"),
                                },
                            },
                        },
                    },
                    
                    // Main container for nixpacks build
                    Containers: []corev1.Container{
                        {
                            Name:  "nixpacks-build",
                            Image: "ghcr.io/railwayapp/nixpacks:latest",
                            Env: []corev1.EnvVar{
                                {
                                    Name:  "DOCKER_CONFIG_JSON",
                                    Value: fmt.Sprintf("{\"%s\"}", registryAuthConfig),
                                },
                            },
                            Command: []string{
                                "sh",
                                "-c",
                                fmt.Sprintf(
                                    "cd /source && nixpacks build . %s -t %s && docker push %s",
                                    generateNixpacksEnvFlags(service.EnvVars),
                                    imageTag,
                                    imageTag,
                                ),
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      sharedVolumeName,
                                    MountPath: "/source",
                                },
                                {
                                    Name:      "docker-socket",
                                    MountPath: "/var/run/docker.sock",
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("1"),
                                    corev1.ResourceMemory: resource.MustParse("1Gi"),
                                },
                            },
                        },
                    },
                    
                    // Volumes
                    Volumes: []corev1.Volume{
                        {
                            Name: sharedVolumeName,
                            VolumeSource: corev1.VolumeSource{
                                EmptyDir: &corev1.EmptyDirVolumeSource{},
                            },
                        },
                        {
                            Name: "docker-socket",
                            VolumeSource: corev1.VolumeSource{
                                HostPath: &corev1.HostPathVolumeSource{
                                    Path: "/var/run/docker.sock",
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    
    return job, nil
}

// Helper function to convert int to *int32
func int32Ptr(i int32) *int32 {
    return &i
}

// Helper to get the appropriate git checkout command
// Returns empty string if CommitSHA is empty (will use HEAD of branch)
func getCheckoutCommand(commitSHA string) string {
    if commitSHA == "" {
        return "" // No checkout needed, will use HEAD of branch
    }
    return "&& git checkout " + commitSHA
}

// generateNixpacksEnvFlags converts a map of environment variables to nixpacks --env flags
// Format: "--env KEY1=VALUE1 --env KEY2=VALUE2 --env KEY3=VALUE3"
func generateNixpacksEnvFlags(envVars models.EnvVars) string {
    if len(envVars) == 0 {
        return ""
    }
    
    var envFlags strings.Builder
    
    // Add each environment variable as a --env flag
    for key, value := range envVars {
        // Properly quote the value to handle special characters
        quotedValue := strings.ReplaceAll(value, "'", "'\"'\"'")
        envFlags.WriteString(fmt.Sprintf("--env %s='%s' ", key, quotedValue))
    }
    
    return envFlags.String()
}

// generateDockerAuthConfig creates the Docker authentication config for pushing to the registry
// Username is always "admin" and password comes from the registry model
func generateDockerAuthConfig(registry models.Registry) string {
    // Create auth string (username:password) and base64 encode it
    authString := fmt.Sprintf("admin:%s", registry.Password)
    encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))
    
    // Return the formatted auth config JSON snippet
    // Use the registry URL directly without http:// prefix
    return fmt.Sprintf("\"auths\": {\"%s\": {\"auth\": \"%s\"}}", registry.URL, encodedAuth)
}

func DeleteBuildResources(service models.Service) error {
	// Delete the job
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	jobName := GetJobName(service.ID, "")
	
	// Delete the job
	err = k8sClient.Clientset.BatchV1().Jobs(GetJobNamespace()).Delete(context.Background(), jobName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete job: %v", err)
	}
	
	return nil
}