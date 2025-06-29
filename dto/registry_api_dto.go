package dto

import "github.com/pendeploy-simple/lib/kubernetes"

// RegistryAPI handles interactions with Docker Registry HTTP API v2 via Kubernetes proxy
type RegistryAPI struct {
	ServiceName string
	Namespace   string
	K8sClient   *kubernetes.Client
}


type TagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type ManifestResponse struct {
	SchemaVersion int             `json:"schemaVersion"`
	MediaType     string          `json:"mediaType"`
	Config        ManifestConfig  `json:"config"`
	Layers        []ManifestLayer `json:"layers"`
	History       []ManifestItem  `json:"history,omitempty"`
}

type ManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type ManifestLayer struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

type ManifestItem struct {
	V1Compatibility string `json:"v1Compatibility"`
}