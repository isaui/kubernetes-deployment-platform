package dto

// CertificateCondition represents a condition of a Kubernetes certificate
type CertificateCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"` 
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
}

// CertificateStats represents processed statistics for a Kubernetes certificate
type CertificateStats struct {
	Name            string                 `json:"name"`
	Namespace       string                 `json:"namespace"`
	Issuer          string                 `json:"issuer,omitempty"`
	SecretName      string                 `json:"secretName"`
	DNSNames        []string               `json:"dnsNames"`
	Status          string                 `json:"status"`
	NotBefore       string                 `json:"notBefore,omitempty"`
	NotAfter        string                 `json:"notAfter,omitempty"`
	RenewalTime     string                 `json:"renewalTime,omitempty"`
	Conditions      []CertificateCondition `json:"conditions,omitempty"`
	IsExpired       bool                   `json:"isExpired"`
	DaysUntilExpiry int                    `json:"daysUntilExpiry"`
	Created         string                 `json:"created"`
}

// CertificateStatsResponse represents the response for certificate statistics API
type CertificateStatsResponse struct {
	Certificates []CertificateStats `json:"certificates"`
}
