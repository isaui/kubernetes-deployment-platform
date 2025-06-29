package utils

import (
	"fmt"
	"log"
	"strings"
	"github.com/pendeploy-simple/models"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createKanikoBuildJob creates a job definition using Kaniko for daemonless builds
func createKanikoBuildJob(registryURL string, deployment models.Deployment, service models.Service, image string) (*batchv1.Job, error) {
    // Create job name based on service and deployment ID
    jobName := GetJobName(service.ID, deployment.ID)
    log.Println("Creating Kaniko job name based on service and deployment ID")
    
    // Parse branch (use default if empty)
    branch := service.Branch
    if branch == "" {
        branch = "main"
    }
    log.Printf("Using branch: %s", branch)
    
    // Sanitize repository URL (ensure it ends with .git)
    repoURL := service.RepoURL
    if !strings.HasSuffix(repoURL, ".git") {
        repoURL = repoURL + ".git"
    }
    log.Printf("Repository URL: %s", repoURL)
    
    // Shared volume for git clone and build
    sharedVolumeName := "build-workspace"
    log.Println("Preparing Kaniko job configuration with optimized resources")
    
    // Create job spec with adequate resources for production builds
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      jobName,
            Namespace: GetJobNamespace(),
            Labels: map[string]string{
                "app":           "pendeploy",
                "service-id":    service.ID,
                "deployment-id": deployment.ID,
                "builder":       "kaniko",
            },
        },
        Spec: batchv1.JobSpec{
            // No retries on failure - immediate feedback
            BackoffLimit: int32Ptr(0),
            
            // Cleanup after 10 minutes (enough time for debugging)
            TTLSecondsAfterFinished: int32Ptr(600),
            
            // Kill job if it runs longer than 20 minutes
            ActiveDeadlineSeconds: int64Ptr(1200),
            
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app":           "pendeploy",
                        "service-id":    service.ID,
                        "deployment-id": deployment.ID,
                        "builder":       "kaniko",
                    },
                },
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    
                    // Init container: Git clone only
                    InitContainers: []corev1.Container{
                        {
                            Name:  "git-clone",
                            Image: "alpine/git:2.43.0", // Public registry
                            Command: []string{
                                "sh",
                                "-c",
                                fmt.Sprintf(
                                    "echo 'Starting git clone...' && git clone --branch %s --single-branch --depth 1 %s /workspace %s && echo 'Git clone completed successfully' && ls -la /workspace && echo 'Checking for Dockerfile...' && ls -la /workspace/Dockerfile",
                                    branch,
                                    repoURL,
                                    getCheckoutCommand(deployment.CommitSHA),
                                ),
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      sharedVolumeName,
                                    MountPath: "/workspace",
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:              resource.MustParse("100m"),
                                    corev1.ResourceMemory:           resource.MustParse("128Mi"),
                                    corev1.ResourceEphemeralStorage: resource.MustParse("512Mi"),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:              resource.MustParse("200m"),
                                    corev1.ResourceMemory:           resource.MustParse("256Mi"),
                                    corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
                                },
                            },
                        },
                    },
                    
                    // Main container: Kaniko executor
                    Containers: []corev1.Container{
                        {
                            Name:  "kaniko-executor",
                            Image: "gcr.io/kaniko-project/executor:v1.23.2", // Public registry
                            Args: append([]string{
                                "--context=/workspace",
                                "--dockerfile=/workspace/Dockerfile",
                                fmt.Sprintf("--destination=%s", image),
                                "--cache=true",
                                fmt.Sprintf("--cache-repo=%s/cache", CleanRegistryURL(registryURL)),
                                "--cache-ttl=168h", // 1 week cache retention
                                "--cleanup",
                                "--verbosity=info",
                                "--log-format=color",
                                "--log-timestamp",
                                "--insecure",
                                "--compressed-caching=false", // Faster processing
                                "--single-snapshot", // Faster layer creation
                            }, generateKanikoBuildArgs(service.EnvVars)...),
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      sharedVolumeName,
                                    MountPath: "/workspace",
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:              resource.MustParse("500m"), // 1 core baseline
                                    corev1.ResourceMemory:           resource.MustParse("1Gi"),   // 2GB baseline for npm + snapshot
                                    corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),   // 4GB for layers + node_modules
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:              resource.MustParse("2000m"), // 2 cores burst capability
                                    corev1.ResourceMemory:           resource.MustParse("6Gi"),   // 6GB max for heavy builds
                                    corev1.ResourceEphemeralStorage: resource.MustParse("12Gi"),  // 12GB max storage
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
                                // Optimize Node.js memory usage during builds
                                {
                                    Name:  "NODE_OPTIONS",
                                    Value: "--max-old-space-size=4096",
                                },
                            },
                        },
                    },
                    
                    // Volumes
                    Volumes: []corev1.Volume{
                        {
                            Name: sharedVolumeName,
                            VolumeSource: corev1.VolumeSource{
                                EmptyDir: &corev1.EmptyDirVolumeSource{
                                    // 8GB should handle most builds comfortably
                                    SizeLimit: resource.NewQuantity(8*1024*1024*1024, resource.BinarySI), // 8GB limit
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    log.Println("Kaniko job spec created successfully with production-ready resources")
    return job, nil
}

// generateKanikoBuildArgs generates --build-arg flags for Kaniko
func generateKanikoBuildArgs(envVars models.EnvVars) []string {
    var buildArgs []string
    
    // Add each environment variable as a build arg
    for key, value := range envVars {
        buildArgs = append(buildArgs, fmt.Sprintf("--build-arg=%s=%s", key, value))
    }
    
    return buildArgs
}