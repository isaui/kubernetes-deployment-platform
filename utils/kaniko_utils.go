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

// createKanikoBuildJob creates a job definition using Kaniko with auto Dockerfile fixing
func createKanikoBuildJob(registryURL string, deployment models.Deployment, service models.Service, image string) (*batchv1.Job, error) {
    jobName := GetJobName(service.ID, deployment.ID)
    log.Println("Creating Kaniko job with Dockerfile auto-fixing")
    
    branch := service.Branch
    if branch == "" {
        branch = "main"
    }
    log.Printf("Using branch: %s", branch)
    
    repoURL := service.RepoURL
    if !strings.HasSuffix(repoURL, ".git") {
        repoURL = repoURL + ".git"
    }
    log.Printf("Repository URL: %s", repoURL)
    
    sharedVolumeName := "build-workspace"
    log.Println("Preparing Kaniko job configuration with Dockerfile auto-fixing")
    
    // Generate Dockerfile fix script
    dockerfileFixScript := generateDockerfileFixScript(service.EnvVars)
    
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
            BackoffLimit: int32Ptr(0),
            TTLSecondsAfterFinished: int32Ptr(600),
            ActiveDeadlineSeconds: int64Ptr(1200),
            
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app":           "pendeploy",
                        "service-id":    service.ID,
                        "deployment-id": deployment.ID,
                        "builder":       "kaniko",
                        "job-name":      jobName, // For log compatibility
                    },
                },
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    
                    InitContainers: []corev1.Container{
                        {
                            Name:  "git-clone",
                            Image: "alpine/git:2.43.0",
                            Command: []string{"sh", "-c"},
                            Args: []string{fmt.Sprintf(`
                                echo "=== Starting git clone ==="
                                git clone --branch %s --single-branch --depth 1 %s /workspace %s
                                cd /workspace
                                echo "Git clone completed successfully"
                                ls -la
                                
                                echo "=== Checking Dockerfile ==="
                                if [ ! -f "Dockerfile" ]; then
                                    echo "ERROR: Dockerfile not found!"
                                    exit 1
                                fi
                                
                                echo "Original Dockerfile:"
                                cat Dockerfile
                                echo "========================="
                                
                                echo "=== Auto-fixing Dockerfile ==="
                                %s
                                
                                echo "Final Dockerfile:"
                                cat Dockerfile
                                echo "================"
                                echo "Dockerfile auto-fixing completed!"
                            `, 
                                branch, 
                                repoURL, 
                                getCheckoutCommand(deployment.CommitSHA),
                                dockerfileFixScript,
                            )},
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
                    
                    Containers: []corev1.Container{
                        {
                            Name:  "kaniko-executor",
                            Image: "gcr.io/kaniko-project/executor:v1.23.2",
                            Args: append([]string{
                                "--context=/workspace",
                                "--dockerfile=/workspace/Dockerfile",
                                fmt.Sprintf("--destination=%s", image),
                                "--cache=true",
                                fmt.Sprintf("--cache-repo=%s/cache", CleanRegistryURL(registryURL)),
                                "--cache-ttl=168h",
                                "--cleanup",
                                "--verbosity=info",
                                "--log-format=color",
                                "--log-timestamp",
                                "--insecure",
                                "--compressed-caching=false",
                                "--single-snapshot",
                            }, generateKanikoBuildArgs(service.EnvVars)...),
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      sharedVolumeName,
                                    MountPath: "/workspace",
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:              resource.MustParse("500m"),
                                    corev1.ResourceMemory:           resource.MustParse("1Gi"),
                                    corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:              resource.MustParse("2000m"),
                                    corev1.ResourceMemory:           resource.MustParse("6Gi"),
                                    corev1.ResourceEphemeralStorage: resource.MustParse("12Gi"),
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
                                {
                                    Name:  "NODE_OPTIONS",
                                    Value: "--max-old-space-size=4096",
                                },
                            },
                        },
                    },
                    
                    Volumes: []corev1.Volume{
                        {
                            Name: sharedVolumeName,
                            VolumeSource: corev1.VolumeSource{
                                EmptyDir: &corev1.EmptyDirVolumeSource{
                                    SizeLimit: resource.NewQuantity(8*1024*1024*1024, resource.BinarySI),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    log.Println("Kaniko job spec created successfully with Dockerfile auto-fixing")
    return job, nil
}

// generateDockerfileFixScript creates shell script to intelligently fix missing ARG/ENV
func generateDockerfileFixScript(envVars models.EnvVars) string {
    var script strings.Builder
    
    script.WriteString(`
                # Create backup of original
                cp Dockerfile Dockerfile.backup
                
                echo "Checking existing ARG/ENV in Dockerfile..."
    `)
    
    // Add specific ARG/ENV from service config - only if missing
    if len(envVars) > 0 {
        script.WriteString("\n                # Check and add missing environment variables\n")
        for key, value := range envVars {
            script.WriteString(fmt.Sprintf(`
                # Check if ARG %s already exists
                if ! grep -q "^ARG %s" Dockerfile; then
                    echo "Adding missing ARG %s"
                    # Find the best place to insert ARG (after FROM line)
                    sed -i '/^FROM /a ARG %s=%s' Dockerfile
                else
                    echo "ARG %s already exists, skipping"
                fi
                
                # Check if ENV %s already exists
                if ! grep -q "^ENV %s" Dockerfile; then
                    echo "Adding missing ENV %s"
                    # Insert ENV after ARG or FROM
                    if grep -q "^ARG " Dockerfile; then
                        sed -i '/^ARG /a ENV %s=$%s' Dockerfile | head -1
                    else
                        sed -i '/^FROM /a ENV %s=%s' Dockerfile
                    fi
                else
                    echo "ENV %s already exists, skipping"
                fi
            `, key, key, key, key, value, key, key, key, key, key, key, key, value, key))
        }
    }
    
    return script.String()
}

// generateKanikoBuildArgs generates --build-arg flags for Kaniko
func generateKanikoBuildArgs(envVars models.EnvVars) []string {
    var buildArgs []string
    
    for key, value := range envVars {
        buildArgs = append(buildArgs, fmt.Sprintf("--build-arg=%s=%s", key, value))
    }
    
    return buildArgs
}

