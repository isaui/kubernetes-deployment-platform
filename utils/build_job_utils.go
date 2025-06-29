package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetJobName(serviceID string, deploymentID string) string {
	log.Println("Creating job name based on deploymentID only")
	return deploymentID
}

// GetJobNamespace returns the namespace for build jobs
func GetJobNamespace() string {
	return "build-and-deploy"
}
// BuildFromGit creates a Kubernetes job with Kaniko and WAITS for completion
// Returns the resulting image URL only after successful build
// FAILS FAST on any error to prevent infinite loops
func BuildFromGit(deployment models.Deployment, service models.Service, registry models.Registry) (string, error) {
    // Define the image tag - use "latest" if commitSHA is empty
    log.Printf("Building image for service: %s, deployment: %s", service.Name, deployment.ID)
    // Use HTTP for local registries, HTTPS for external ones
    cleanRegistryURL := CleanRegistryURL(registry.URL)
    
    var protocol string
    if strings.Contains(cleanRegistryURL, "local") || 
       strings.Contains(cleanRegistryURL, "localhost") || 
       strings.Contains(cleanRegistryURL, "127.0.0.1") || 
       strings.Contains(cleanRegistryURL, "svc.cluster.local") {
        // Use HTTP for local registries and Kubernetes internal registries
        protocol = "http://"
        log.Printf("Using HTTP protocol for local/cluster registry: %s", cleanRegistryURL)
    } else {
        // Use HTTPS for external registries
        protocol = "https://"
        log.Printf("Using HTTPS protocol for external registry: %s", cleanRegistryURL)
    }
    
    registryURL := fmt.Sprintf("%s%s", protocol, cleanRegistryURL)
    
    image := GenerateImage(cleanRegistryURL, service, deployment)
    log.Printf("Target image (for K8s): %s", image)
    log.Printf("Registry URL (for kaniko): %s", registryURL)
    
    // Create Kubernetes client
    k8sClient, err := kubernetes.NewClient()
    if err != nil {
        log.Printf("FATAL: Failed to create Kubernetes client: %v", err)
        return "", fmt.Errorf("kubernetes client creation failed: %v", err)
    }
    log.Println("Kubernetes client created successfully")
    
    // Ensure namespace exists
    namespace := GetJobNamespace()
    err = EnsureNamespaceExists(namespace)
    if err != nil {
        log.Printf("FATAL: Failed to ensure namespace %s exists: %v", namespace, err)
        return "", fmt.Errorf("namespace creation failed: %v", err)
    }
    log.Printf("Namespace %s confirmed", namespace)
    
    // Cleanup any existing job with the same name first
    jobName := GetJobName(service.ID, deployment.ID)
    log.Printf("Cleaning up existing job: %s", jobName)
    err = cleanupExistingJob(k8sClient, jobName, namespace)
    if err != nil {
        log.Printf("WARNING: Failed to cleanup existing job %s: %v", jobName, err)
        // Continue anyway - this shouldn't be fatal
    }
    
    log.Printf("Creating Kaniko job: %s", jobName)
    // Create the job - pass all necessary parameters
    job, err := createKanikoBuildJob(registryURL, deployment, service, image)
    if err != nil {
        log.Printf("FATAL: Failed to create job definition: %v", err)
        return "", fmt.Errorf("job definition creation failed: %v", err)
    }
    log.Println("Kaniko job definition created successfully")
    
    // Submit the job to Kubernetes
    log.Printf("Submitting job %s to Kubernetes", jobName)
    _, err = k8sClient.Clientset.BatchV1().Jobs(GetJobNamespace()).Create(
        context.Background(),
        job,
        metav1.CreateOptions{},
    )
    
    if err != nil {
        log.Printf("FATAL: Failed to submit job to Kubernetes: %v", err)
        return "", fmt.Errorf("job submission failed: %v", err)
    }
    log.Printf("Job %s submitted to Kubernetes successfully", jobName)
    
    // Wait for build completion with 12-MINUTE timeout
    log.Printf("Waiting for build job %s to complete (timeout: 12 minutes)...", jobName)
    err = waitForJobCompletion(k8sClient, jobName, namespace, 12*time.Minute)
    if err != nil {
        log.Printf("BUILD FAILED: Job %s failed with error: %v", jobName, err)
        
        // Try to get more detailed logs for debugging
        detailedLogs := getDetailedJobLogs(k8sClient, jobName, namespace)
        if detailedLogs != "" {
            log.Printf("Detailed failure logs for job %s:\n%s", jobName, detailedLogs)
        }
        
        return "", fmt.Errorf("build job failed: %v", err)
    }
    
    log.Printf("BUILD SUCCESS: Job %s completed successfully! Image ready: %s", jobName, image)
    return image, nil
}
// waitForJobCompletion waits for a Kubernetes job to complete successfully
// Uses WATCH API for real-time updates - no polling!
func waitForJobCompletion(k8sClient *kubernetes.Client, jobName, namespace string, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    log.Printf("Starting REAL-TIME job monitoring for: %s (timeout: %v)", jobName, timeout)
    
    // Start watching job status changes
    jobWatch, err := k8sClient.Clientset.BatchV1().Jobs(namespace).Watch(ctx, metav1.ListOptions{
        FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
    })
    if err != nil {
        return fmt.Errorf("failed to start job watch: %v", err)
    }
    defer jobWatch.Stop()
    
    // Start watching pod status changes  
    podWatch, err := k8sClient.Clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
        LabelSelector: fmt.Sprintf("job-name=%s", jobName),
    })
    if err != nil {
        return fmt.Errorf("failed to start pod watch: %v", err)
    }
    defer podWatch.Stop()
    
    // Monitor both job and pod events simultaneously
    for {
        select {
        case <-ctx.Done():
            log.Printf("TIMEOUT: Job %s did not complete within %v", jobName, timeout)
            return fmt.Errorf("timeout waiting for job %s to complete (waited %v)", jobName, timeout)
            
        case jobEvent := <-jobWatch.ResultChan():
            if jobEvent.Object == nil {
                log.Printf("Job watch closed, job may have been deleted")
                return fmt.Errorf("job %s watch closed unexpectedly", jobName)
            }
            
            job, ok := jobEvent.Object.(*batchv1.Job)
            if !ok {
                continue
            }
            
            log.Printf("Job %s event: %s", jobName, jobEvent.Type)
            
            // Check job conditions
            for _, condition := range job.Status.Conditions {
                switch condition.Type {
                case batchv1.JobComplete:
                    if condition.Status == corev1.ConditionTrue {
                        log.Printf("🎉 SUCCESS: Job %s completed successfully", jobName)
                        return nil
                    }
                    
                case batchv1.JobFailed:
                    if condition.Status == corev1.ConditionTrue {
                        reason := "Unknown failure"
                        if condition.Reason != "" {
                            reason = condition.Reason
                        }
                        if condition.Message != "" {
                            reason = fmt.Sprintf("%s: %s", reason, condition.Message)
                        }
                        
                        log.Printf("❌ FAILED: Job %s failed with reason: %s", jobName, reason)
                        return fmt.Errorf("job failed: %s", reason)
                    }
                }
            }
            
            // Log status changes
            active := job.Status.Active
            succeeded := job.Status.Succeeded
            failed := job.Status.Failed
            log.Printf("Job %s status update - Active: %d, Succeeded: %d, Failed: %d", 
                jobName, active, succeeded, failed)
                
        case podEvent := <-podWatch.ResultChan():
            if podEvent.Object == nil {
                continue
            }
            
            pod, ok := podEvent.Object.(*corev1.Pod)
            if !ok {
                continue
            }
            
            log.Printf("Pod %s event: %s, phase: %s", pod.Name, podEvent.Type, pod.Status.Phase)
            
            // Check for immediate pod failures
            podError := checkPodForErrors(pod)
            if podError != nil {
                log.Printf("🚨 POD ERROR: Pod %s has error: %v", pod.Name, podError)
                return fmt.Errorf("pod error in job %s: %v", jobName, podError)
            }
        }
    }
}

// checkPodForErrors checks a single pod for error conditions
func checkPodForErrors(pod *corev1.Pod) error {
    // Check pod phase for immediate failures
    switch pod.Status.Phase {
    case corev1.PodFailed:
        return fmt.Errorf("pod %s failed with reason: %s", pod.Name, pod.Status.Reason)
    }
    
    // Check container statuses for ImagePullBackOff and other waiting issues
    for _, containerStatus := range pod.Status.ContainerStatuses {
        if containerStatus.State.Waiting != nil {
            waiting := containerStatus.State.Waiting
            switch waiting.Reason {
            case "ImagePullBackOff", "ErrImagePull", "InvalidImageName":
                return fmt.Errorf("container %s image pull failed: %s - %s", 
                    containerStatus.Name, waiting.Reason, waiting.Message)
            case "CreateContainerConfigError", "CreateContainerError":
                return fmt.Errorf("container %s config error: %s - %s", 
                    containerStatus.Name, waiting.Reason, waiting.Message)
            case "CrashLoopBackOff":
                return fmt.Errorf("container %s crash loop: %s - %s", 
                    containerStatus.Name, waiting.Reason, waiting.Message)
            }
        }
        
        // Check for containers that have been restarting too much
        if containerStatus.RestartCount > 3 {
            return fmt.Errorf("container %s restarted too many times (%d)", 
                containerStatus.Name, containerStatus.RestartCount)
        }
    }
    
    // Check init container statuses
    for _, initStatus := range pod.Status.InitContainerStatuses {
        if initStatus.State.Waiting != nil {
            waiting := initStatus.State.Waiting
            switch waiting.Reason {
            case "ImagePullBackOff", "ErrImagePull", "InvalidImageName":
                return fmt.Errorf("init container %s image pull failed: %s - %s", 
                    initStatus.Name, waiting.Reason, waiting.Message)
            case "CreateContainerConfigError", "CreateContainerError":
                return fmt.Errorf("init container %s config error: %s - %s", 
                    initStatus.Name, waiting.Reason, waiting.Message)
            }
        }
    }
    
    return nil
}



