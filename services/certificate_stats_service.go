package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pendeploy-simple/dto"
	"github.com/pendeploy-simple/lib/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CertificateStatsService handles operations related to certificate resources
type CertificateStatsService struct{}

// NewCertificateStatsService creates a new certificate stats service
func NewCertificateStatsService() *CertificateStatsService {
	return &CertificateStatsService{}
}

// GetCertificateStats returns statistics about cert-manager certificates in the specified namespace
func (s *CertificateStatsService) GetCertificateStats(namespace string) (dto.CertificateStatsResponse, error) {
	ctx := context.Background()

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return dto.CertificateStatsResponse{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	// Use dynamic client to get certificates
	dynamic := kubeClient.Clientset.Discovery().RESTClient()
	var certList *unstructured.UnstructuredList

	// Try to get certificates using dynamic client
	result, err := dynamic.Get().
		AbsPath("/apis/cert-manager.io/v1/").
		Namespace(namespace).
		Resource("certificates").
		DoRaw(ctx)

	if err != nil {
		// Return empty response if cert-manager is not installed
		return dto.CertificateStatsResponse{
			Certificates: []dto.CertificateStats{},
		}, nil
	}

	// Parse response
	certList = &unstructured.UnstructuredList{}
	if err := certList.UnmarshalJSON(result); err != nil {
		return dto.CertificateStatsResponse{}, fmt.Errorf("failed to parse certificates: %v", err)
	}

	// Process certificates
	certStats := make([]dto.CertificateStats, 0)
	for _, item := range certList.Items {
		// Extract basic metadata
		name := item.GetName()
		ns := item.GetNamespace()
		created := item.GetCreationTimestamp().Format(time.RFC3339)

		// Extract spec fields
		spec, found, err := unstructured.NestedMap(item.Object, "spec")
		if err != nil || !found {
			continue
		}

		// Get secretName
		secretName, _ := spec["secretName"].(string)

		// Get issuer
		issuerRef, found, _ := unstructured.NestedMap(item.Object, "spec", "issuerRef")
		issuer := ""
		if found {
			issuerKind, _ := issuerRef["kind"].(string)
			issuerName, _ := issuerRef["name"].(string)
			issuer = fmt.Sprintf("%s/%s", issuerKind, issuerName)
		}

		// Extract DNS names
		dnsNames := []string{}
		if dnsNamesIface, found := spec["dnsNames"]; found {
			if dnsNamesArr, ok := dnsNamesIface.([]interface{}); ok {
				for _, dnsi := range dnsNamesArr {
					if dns, ok := dnsi.(string); ok {
						dnsNames = append(dnsNames, dns)
					}
				}
			}
		}

		// Extract status
		status := "Unknown"
		notBefore := ""
		notAfter := ""
		renewalTime := ""
		isExpired := false
		daysUntilExpiry := 0

		if _, found, _ := unstructured.NestedMap(item.Object, "status"); found {
			if conditions, found, _ := unstructured.NestedSlice(item.Object, "status", "conditions"); found && len(conditions) > 0 {
				// Get the most recent condition (last one)
				lastCondition, ok := conditions[len(conditions)-1].(map[string]interface{})
				if ok {
					condType, _ := lastCondition["type"].(string)
					condStatus, _ := lastCondition["status"].(string)
					if condType == "Ready" && condStatus == "True" {
						status = "Ready"
					} else {
						reason, _ := lastCondition["reason"].(string)
						if reason != "" {
							status = reason
						}
					}
				}
			}

			// Check if certificate is issued and get expiry dates
			if notBeforeStr, found, _ := unstructured.NestedString(item.Object, "status", "notBefore"); found {
				notBefore = notBeforeStr
			}
			
			if notAfterStr, found, _ := unstructured.NestedString(item.Object, "status", "notAfter"); found {
				notAfter = notAfterStr
				
				// Try to parse the expiry date
				if expiryTime, err := time.Parse(time.RFC3339, notAfter); err == nil {
					now := time.Now()
					isExpired = now.After(expiryTime)
					daysUntilExpiry = int(expiryTime.Sub(now).Hours() / 24)
					if daysUntilExpiry < 0 {
						daysUntilExpiry = 0
					}
				}
			}
			
			if renewalTimeStr, found, _ := unstructured.NestedString(item.Object, "status", "renewalTime"); found {
				renewalTime = renewalTimeStr
			}
		}

		// Extract conditions
		conditions := []dto.CertificateCondition{}
		if conditionsIface, found, _ := unstructured.NestedSlice(item.Object, "status", "conditions"); found {
			for _, ci := range conditionsIface {
				if cond, ok := ci.(map[string]interface{}); ok {
					condType, _ := cond["type"].(string)
					condStatus, _ := cond["status"].(string)
					lastTransitionTime, _ := cond["lastTransitionTime"].(string)
					reason, _ := cond["reason"].(string)
					message, _ := cond["message"].(string)
					
					conditions = append(conditions, dto.CertificateCondition{
						Type:               condType,
						Status:             condStatus,
						LastTransitionTime: lastTransitionTime,
						Reason:             reason,
						Message:            message,
					})
				}
			}
		}

		// Add certificate stats
		certStats = append(certStats, dto.CertificateStats{
			Name:            name,
			Namespace:       ns,
			Issuer:          issuer,
			SecretName:      secretName,
			DNSNames:        dnsNames,
			Status:          status,
			NotBefore:       notBefore,
			NotAfter:        notAfter,
			RenewalTime:     renewalTime,
			Conditions:      conditions,
			IsExpired:       isExpired,
			DaysUntilExpiry: daysUntilExpiry,
			Created:         created,
		})
	}

	return dto.CertificateStatsResponse{
		Certificates: certStats,
	}, nil
}
