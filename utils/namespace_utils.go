package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/pendeploy-simple/lib/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureNamespaceExists checks if a namespace exists and creates it if it doesn't
func EnsureNamespaceExists(namespaceName string) error {
	log.Println("Ensuring namespace exists:", namespaceName)
	
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	
	// Check if namespace exists
	_, err = k8sClient.Clientset.CoreV1().Namespaces().Get(
		context.Background(),
		namespaceName,
		metav1.GetOptions{},
	)
	
	// If namespace exists, return
	if err == nil {
		log.Println("Namespace already exists:", namespaceName)
		return nil
	}
	
	// If error is not "not found", return error
	if !errors.IsNotFound(err) {
		return fmt.Errorf("error checking namespace: %v", err)
	}
	
	// Create namespace
	log.Println("Creating namespace:", namespaceName)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	
	_, err = k8sClient.Clientset.CoreV1().Namespaces().Create(
		context.Background(),
		namespace,
		metav1.CreateOptions{},
	)
	
	if err != nil {
		return fmt.Errorf("error creating namespace: %v", err)
	}
	
	log.Println("Namespace created successfully:", namespaceName)
	return nil
}


// EnsureNamespaceExistsWithCA creates namespace if it doesn't exist and replicates CA ConfigMap
func EnsureNamespaceExistsWithCA(namespace string, registryNamespace string, caConfigMapName string) error {
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	ctx := context.Background()

	// Ensure the namespace exists
	_, err = k8sClient.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the namespace if it doesn't exist
			_, err = k8sClient.Clientset.CoreV1().Namespaces().Create(
				ctx,
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespace,
						Labels: map[string]string{
							"managed-by": "pendeploy",
						},
					},
				},
				metav1.CreateOptions{},
			)
			if err != nil {
				return fmt.Errorf("failed to create namespace: %v", err)
			}
			log.Printf("Namespace %s created successfully", namespace)
		} else {
			return fmt.Errorf("failed to get namespace: %v", err)
		}
	}

	// Replicate CA ConfigMap to this namespace for future registry operations
	err = replicateCAConfigMapToNamespace(namespace, registryNamespace, caConfigMapName)
	if err != nil {
		// Log warning but don't fail deployment - CA replication is optional
		log.Printf("Warning: Failed to replicate CA ConfigMap to namespace %s: %v", namespace, err)
	}

	return nil
}

// replicateCAConfigMapToNamespace replicates the global CA ConfigMap to the target namespace
func replicateCAConfigMapToNamespace(targetNamespace string, registryNamespace string, caConfigMapName string) error {
	k8sClient, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Check if CA ConfigMap exists in registry namespace
	globalCA, err := k8sClient.Clientset.CoreV1().ConfigMaps(registryNamespace).Get(
		context.Background(), 
		caConfigMapName, 
		metav1.GetOptions{},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			// CA doesn't exist yet, skip replication
			log.Printf("Global CA ConfigMap not found, skipping replication to namespace %s", targetNamespace)
			return nil
		}
		return fmt.Errorf("failed to get global CA ConfigMap: %v", err)
	}

	// Create replicated ConfigMap in target namespace
	replicatedCA := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caConfigMapName,
			Namespace: targetNamespace,
			Labels: map[string]string{
				"app":                 "pendeploy",
				"type":               "registry-ca",
				"replicated-from":    registryNamespace,
			},
		},
		Data: globalCA.Data,
	}

	// Try to create or update
	_, err = k8sClient.Clientset.CoreV1().ConfigMaps(targetNamespace).Create(
		context.Background(), 
		replicatedCA, 
		metav1.CreateOptions{},
	)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Update existing ConfigMap
			_, err = k8sClient.Clientset.CoreV1().ConfigMaps(targetNamespace).Update(
				context.Background(), 
				replicatedCA, 
				metav1.UpdateOptions{},
			)
			if err != nil {
				return fmt.Errorf("failed to update CA ConfigMap in namespace %s: %v", targetNamespace, err)
			}
			log.Printf("CA ConfigMap updated in namespace %s", targetNamespace)
		} else {
			return fmt.Errorf("failed to create CA ConfigMap in namespace %s: %v", targetNamespace, err)
		}
	} else {
		log.Printf("CA ConfigMap replicated to namespace %s", targetNamespace)
	}

	return nil
}
