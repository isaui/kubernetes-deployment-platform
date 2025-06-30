package kubernetes

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// DeleteNamespace deletes a kubernetes namespace and all resources within it
func (c *Client) DeleteNamespace(namespace string) error {
	// Create a context with timeout
	ctx := context.Background()

	// Delete the namespace which will cascade delete all resources in it
	err := c.Clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{
		PropagationPolicy: func() *metav1.DeletionPropagation {
			policy := metav1.DeletePropagationForeground
			return &policy
		}(),
	})

	if err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", namespace, err)
	}

	return nil
}

// NamespaceExists checks if a namespace exists in the cluster
func (c *Client) NamespaceExists(namespace string) (bool, error) {
	ctx := context.Background()
	
	_, err := c.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		// Check if error is because namespace doesn't exist
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("error checking namespace %s: %w", namespace, err)
	}
	
	return true, nil
}
