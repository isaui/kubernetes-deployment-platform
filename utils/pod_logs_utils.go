package utils

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"github.com/pendeploy-simple/lib/kubernetes"
)

// getDetailedJobLogs retrieves comprehensive logs from all containers in a job's pods
func getDetailedJobLogs(k8sClient *kubernetes.Client, jobName, namespace string) string {
    log.Printf("Retrieving detailed logs for failed job: %s", jobName)
    
    // Find pods for this job
    pods, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(
        context.Background(),
        metav1.ListOptions{
            LabelSelector: fmt.Sprintf("job-name=%s", jobName),
        },
    )
    
    if err != nil || len(pods.Items) == 0 {
        return fmt.Sprintf("No pods found for job %s: %v", jobName, err)
    }
    
    var allLogs strings.Builder
    
    for _, pod := range pods.Items {
        allLogs.WriteString(fmt.Sprintf("\n=== Pod: %s ===\n", pod.Name))
        
        // Get logs from all containers
        containers := []string{"git-clone", "dockerfile-generator", "kaniko-executor"}
        
        for _, container := range containers {
            allLogs.WriteString(fmt.Sprintf("\n--- Container: %s ---\n", container))
            
            req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
                Container: container,
                TailLines: int64Ptr(20), // Last 20 lines per container
            })
            
            logs, err := req.Stream(context.Background())
            if err != nil {
                allLogs.WriteString(fmt.Sprintf("Error getting logs for container %s: %v\n", container, err))
                continue
            }
            
            // Read logs with size limit
            buffer := make([]byte, 2048) // 2KB per container
            n, _ := logs.Read(buffer)
            logs.Close()
            
            if n > 0 {
                allLogs.WriteString(string(buffer[:n]))
            } else {
                allLogs.WriteString("No logs available\n")
            }
        }
    }
    
    return allLogs.String()
}

// getJobPodLogs retrieves the last few lines of logs from a job's pod for debugging
func getJobPodLogs(k8sClient *kubernetes.Client, jobName, namespace string) string {
    // Find pods for this job
    pods, err := k8sClient.Clientset.CoreV1().Pods(namespace).List(
        context.Background(),
        metav1.ListOptions{
            LabelSelector: fmt.Sprintf("job-name=%s", jobName),
        },
    )
    
    if err != nil || len(pods.Items) == 0 {
        return "No pod logs available"
    }
    
    // Get logs from the main container of the first pod
    podName := pods.Items[0].Name
    
    // Try to get logs from kaniko-executor container
    req := k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
        Container: "kaniko-executor",
        TailLines: int64Ptr(10), // Last 10 lines
    })
    
    logs, err := req.Stream(context.Background())
    if err != nil {
        return fmt.Sprintf("Error getting logs: %v", err)
    }
    defer logs.Close()
    
    // Read logs with size limit
    buffer := make([]byte, 1024) // 1KB limit
    n, _ := logs.Read(buffer)
    
    return string(buffer[:n])
}
