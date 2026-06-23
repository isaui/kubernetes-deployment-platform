package kubernetes

import (
	"fmt"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ProxyOptions contains options for connecting through a local Kubernetes API proxy.
type ProxyOptions struct {
	// Host is the proxy URL.
	Host string
}

// Client represents a kubernetes client
type Client struct {
	Clientset     *kubernetes.Clientset
	MetricsClient *metricsv1beta1.Clientset
	DynamicClient dynamic.Interface
}

// NewClient creates a Kubernetes client.
// If K8S_PROXY_URL is set, it is used for local development. Otherwise the
// client uses in-cluster ServiceAccount credentials.
func NewClient() (*Client, error) {
	proxyURL := os.Getenv("K8S_PROXY_URL")
	if proxyURL != "" {
		return NewClientWithOptions(ProxyOptions{Host: proxyURL})
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster Kubernetes config: %v", err)
	}

	return NewClientWithConfig(config)
}

// NewClientWithOptions creates a new Kubernetes client with the specified proxy options.
func NewClientWithOptions(options ProxyOptions) (*Client, error) {
	host := options.Host
	if host == "" {
		return nil, fmt.Errorf("proxy host is required")
	}

	config := &rest.Config{
		Host: host,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	return NewClientWithConfig(config)
}

// NewClientWithConfig creates Kubernetes clients from a rest.Config.
func NewClientWithConfig(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	metricsClient, err := metricsv1beta1.NewForConfig(config)
	if err != nil {
		// If metrics client fails, we'll continue without it
		fmt.Printf("Warning: Unable to create metrics client: %v\n", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		// If dynamic client fails, return error as it's needed for custom resources
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &Client{
		Clientset:     clientset,
		MetricsClient: metricsClient,
		DynamicClient: dynamicClient,
	}, nil
}

// GetConfig returns a Kubernetes REST config.
// If K8S_PROXY_URL is set, it is used for local development. Otherwise the
// config uses in-cluster ServiceAccount credentials.
func GetConfig() (*rest.Config, error) {
	proxyURL := os.Getenv("K8S_PROXY_URL")
	if proxyURL != "" {
		return GetConfigWithHost(proxyURL)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster Kubernetes config: %v", err)
	}

	return config, nil
}

// GetConfigWithHost returns a Kubernetes config using the specified API proxy host.
func GetConfigWithHost(host string) (*rest.Config, error) {
	if host == "" {
		return nil, fmt.Errorf("proxy host is required")
	}

	return &rest.Config{
		Host: host,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}, nil
}
