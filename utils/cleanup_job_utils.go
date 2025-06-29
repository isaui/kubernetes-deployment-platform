package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// cleanupJobsForService removes all jobs associated with a service
func cleanupJobsForService(k8sClient *kubernetes.Client, serviceID string) error {
    namespace := GetJobNamespace()
    
    // List all jobs with the service label
    jobs, err := k8sClient.Clientset.BatchV1().Jobs(namespace).List(
        context.Background(),
        metav1.ListOptions{
            LabelSelector: fmt.Sprintf("service-id=%s", serviceID),
        },
    )
    
    if err != nil {
        return fmt.Errorf("failed to list jobs for service %s: %v", serviceID, err)
    }
    
    // Delete each job
    for _, job := range jobs.Items {
        err = k8sClient.Clientset.BatchV1().Jobs(namespace).Delete(
            context.Background(),
            job.Name,
            metav1.DeleteOptions{
                PropagationPolicy: &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
            },
        )
        
        if err != nil && !errors.IsNotFound(err) {
            log.Printf("Failed to delete job %s: %v", job.Name, err)
        } else {
            log.Printf("Successfully deleted job %s", job.Name)
        }
    }
    
    return nil
}

// cleanupExistingJob removes any existing job with the same name
func cleanupExistingJob(k8sClient *kubernetes.Client, jobName, namespace string) error {
    // Check if job exists first to avoid unnecessary delete calls
    _, err := k8sClient.Clientset.BatchV1().Jobs(namespace).Get(
        context.Background(),
        jobName,
        metav1.GetOptions{},
    )
    
    // If job doesn't exist, no need to delete
    if errors.IsNotFound(err) {
        log.Printf("Job %s not found, no cleanup needed", jobName)
        return nil
    }
    
    // If error checking job (other than not found), return error
    if err != nil {
        return fmt.Errorf("failed to check existing job: %v", err)
    }
    
    // Job exists, proceed with deletion
    log.Printf("Job %s exists, cleaning up", jobName)
    err = k8sClient.Clientset.BatchV1().Jobs(namespace).Delete(
        context.Background(),
        jobName,
        metav1.DeleteOptions{
            PropagationPolicy: &[]metav1.DeletionPropagation{metav1.DeletePropagationBackground}[0],
        },
    )
    
    if err != nil {
        return fmt.Errorf("failed to delete existing job: %v", err)
    }
    
    log.Printf("Job %s cleaned up successfully", jobName)
    return nil
}

// DeleteBuildResources cleans up build-related resources for a service
func DeleteBuildResources(service models.Service) error {
    k8sClient, err := kubernetes.NewClient()
    if err != nil {
        return fmt.Errorf("failed to create Kubernetes client: %v", err)
    }
    
    // Clean up all jobs for this service
    err = cleanupJobsForService(k8sClient, service.ID)
    if err != nil {
        log.Println("Warning: failed to cleanup jobs for service:", err)
    }
    
    return nil
}
