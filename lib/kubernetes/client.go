package kubernetes

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ProxyOptions contains options for connecting to kubectl proxy
type ProxyOptions struct {
	// Host is the kubectl proxy URL (default: http://localhost:8001)
	Host string
}

// Client represents a kubernetes client
type Client struct {
	Clientset     *kubernetes.Clientset
	MetricsClient *metricsv1beta1.Clientset
}

// NewClient creates a new Kubernetes client using the proxy address from env or default
func NewClient() (*Client, error) {
	// Read from environment or use default
	proxyURL := os.Getenv("K8S_PROXY_URL")
	if proxyURL == "" {
		proxyURL = "http://localhost:8001"
	}
	
	return NewClientWithOptions(ProxyOptions{
		Host: proxyURL,
	})
}

// NewClientWithOptions creates a new Kubernetes client with the specified proxy options
func NewClientWithOptions(options ProxyOptions) (*Client, error) {
	// Set default proxy host if not provided
	host := options.Host
	if host == "" {
		host = "http://localhost:8001"
	}

	// Create a simple REST config pointing to the kubectl proxy
	config := &rest.Config{
		Host: host,
		// No authentication needed when using kubectl proxy
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Create metrics client
	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		// If metrics client fails, we'll continue without it
		fmt.Printf("Warning: Unable to create metrics client: %v\n", err)
	}

	return &Client{
		Clientset:     clientset,
		MetricsClient: metricsClient,
	}, nil
}

// GetConfig returns a REST config for the proxy address from env or default
func GetConfig() (*rest.Config, error) {
	// Read from environment or use default
	proxyURL := os.Getenv("K8S_PROXY_URL")
	if proxyURL == "" {
		proxyURL = "http://localhost:8001"
	}
	
	return GetConfigWithHost(proxyURL)
}

// GetConfigWithHost returns a Kubernetes config using the specified kubectl proxy host
func GetConfigWithHost(host string) (*rest.Config, error) {
	if host == "" {
		host = "http://localhost:8001"
	}
	
	return &rest.Config{
		Host: host,
		// No authentication needed when using kubectl proxy
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}, nil
}
