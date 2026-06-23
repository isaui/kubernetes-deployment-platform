package utils

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pendeploy-simple/lib/kubernetes"
	"github.com/pendeploy-simple/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultTCPProxyNamespace = "kubesa-system"
	defaultTCPProxyName      = "tcp-proxy"
	defaultTCPProxyPortStart = 24000
	defaultTCPProxyPortEnd   = 24999
)

type TCPProxyConfig struct {
	Host      string
	Namespace string
	Name      string
	PortStart int
	PortEnd   int
}

func GetTCPProxyConfig() TCPProxyConfig {
	host := strings.TrimSpace(os.Getenv("TCP_PROXY_HOST"))
	if host == "" {
		host = fmt.Sprintf("proxy.%s", GetDefaultDomain())
	}

	return TCPProxyConfig{
		Host:      host,
		Namespace: getEnvString("TCP_PROXY_NAMESPACE", defaultTCPProxyNamespace),
		Name:      getEnvString("TCP_PROXY_NAME", defaultTCPProxyName),
		PortStart: getEnvInt("TCP_PROXY_PORT_START", defaultTCPProxyPortStart),
		PortEnd:   getEnvInt("TCP_PROXY_PORT_END", defaultTCPProxyPortEnd),
	}
}

func EnsureTCPProxyExists(services []models.Service) error {
	client, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	cfg := GetTCPProxyConfig()
	if cfg.PortEnd < cfg.PortStart {
		return fmt.Errorf("invalid TCP proxy port range: %d-%d", cfg.PortStart, cfg.PortEnd)
	}

	ctx := context.Background()
	if err := EnsureNamespaceExists(cfg.Namespace); err != nil {
		return fmt.Errorf("failed to ensure TCP proxy namespace: %w", err)
	}

	configMap := createTCPProxyConfigMap(cfg, services)
	if err := applyTCPProxyConfigMap(ctx, client, configMap); err != nil {
		return err
	}

	deployment := createTCPProxyDeployment(cfg)
	if err := applyTCPProxyDeployment(ctx, client, deployment); err != nil {
		return err
	}

	service := createTCPProxyService(cfg, services)
	if service == nil {
		return deleteTCPProxyService(ctx, client, cfg)
	}

	return applyTCPProxyService(ctx, client, service)
}

func createTCPProxyConfigMap(cfg TCPProxyConfig, services []models.Service) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    map[string]string{"app": cfg.Name},
		},
		Data: map[string]string{
			"haproxy.cfg": buildHAProxyConfig(services),
		},
	}
}

func buildHAProxyConfig(services []models.Service) string {
	var b strings.Builder
	b.WriteString(`global
  log stdout format raw local0
  maxconn 4096

defaults
  log global
  mode tcp
  option tcplog
  timeout connect 10s
  timeout client 1h
  timeout server 1h

`)

	sort.Slice(services, func(i, j int) bool {
		return services[i].ExternalPort < services[j].ExternalPort
	})

	for _, service := range services {
		if !isTCPProxyService(service) {
			continue
		}

		resourceName := GetResourceName(service)
		frontendName := fmt.Sprintf("svc_%s_%d", strings.ReplaceAll(service.ID, "-", "_"), service.ExternalPort)
		backendName := fmt.Sprintf("backend_%s", strings.ReplaceAll(service.ID, "-", "_"))
		targetHost := fmt.Sprintf("%s.%s.svc.cluster.local", resourceName, service.EnvironmentID)

		b.WriteString(fmt.Sprintf("frontend %s\n", frontendName))
		b.WriteString(fmt.Sprintf("  bind *:%d\n", service.ExternalPort))
		b.WriteString(fmt.Sprintf("  default_backend %s\n\n", backendName))
		b.WriteString(fmt.Sprintf("backend %s\n", backendName))
		b.WriteString(fmt.Sprintf("  server primary %s:%d check\n\n", targetHost, service.Port))
	}

	return b.String()
}

func createTCPProxyDeployment(cfg TCPProxyConfig) *appsv1.Deployment {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    map[string]string{"app": cfg.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": cfg.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": cfg.Name},
					Annotations: map[string]string{
						"kubesa.io/restarted-at": time.Now().Format(time.RFC3339),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "haproxy",
							Image: "haproxy:2.9-alpine",
							Args:  []string{"-f", "/usr/local/etc/haproxy/haproxy.cfg"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/usr/local/etc/haproxy",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: cfg.Name},
								},
							},
						},
					},
				},
			},
		},
	}

	SecurePodSpec(&deployment.Spec.Template.Spec)
	return deployment
}

func createTCPProxyService(cfg TCPProxyConfig, services []models.Service) *corev1.Service {
	ports := make([]corev1.ServicePort, 0)
	seenPorts := make(map[int]bool)

	for _, service := range services {
		if !isTCPProxyService(service) || seenPorts[service.ExternalPort] {
			continue
		}
		seenPorts[service.ExternalPort] = true
		ports = append(ports, corev1.ServicePort{
			Name:       fmt.Sprintf("tcp-%d", service.ExternalPort),
			Port:       int32(service.ExternalPort),
			TargetPort: intstr.FromInt(service.ExternalPort),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	if len(ports) == 0 {
		return nil
	}

	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    map[string]string{"app": cfg.Name},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{"app": cfg.Name},
			Ports:    ports,
		},
	}
}

func isTCPProxyService(service models.Service) bool {
	return service.Type == models.ServiceTypeManaged &&
		service.ExternalPort > 0 &&
		service.EnvironmentID != "" &&
		service.Port > 0
}

func applyTCPProxyConfigMap(ctx context.Context, client *kubernetes.Client, configMap *corev1.ConfigMap) error {
	_, err := client.Clientset.CoreV1().ConfigMaps(configMap.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = client.Clientset.CoreV1().ConfigMaps(configMap.Namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	}
	return err
}

func applyTCPProxyDeployment(ctx context.Context, client *kubernetes.Client, deployment *appsv1.Deployment) error {
	_, err := client.Clientset.AppsV1().Deployments(deployment.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = client.Clientset.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	}
	return err
}

func applyTCPProxyService(ctx context.Context, client *kubernetes.Client, service *corev1.Service) error {
	_, err := client.Clientset.CoreV1().Services(service.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = client.Clientset.CoreV1().Services(service.Namespace).Update(ctx, service, metav1.UpdateOptions{})
	}
	return err
}

func deleteTCPProxyService(ctx context.Context, client *kubernetes.Client, cfg TCPProxyConfig) error {
	err := client.Clientset.CoreV1().Services(cfg.Namespace).Delete(ctx, cfg.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func getEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
