package utils

import (
	"log"
	"strings"
	"github.com/pendeploy-simple/models"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// createKpackBuildImage creates a kpack Image resource for Cloud Native Buildpacks builds
func createKpackBuildImage(registryURL string, deployment models.Deployment, service models.Service, image string) (*unstructured.Unstructured, error) {
	imageName := GetJobName(service.ID, deployment.ID)
	log.Println("Creating kpack Image name based on service and deployment ID")
	
	branch := service.Branch
	if branch == "" {
		branch = "main"
	}
	log.Printf("Using branch: %s", branch)
	
	repoURL := service.RepoURL
	if !strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL + ".git"
	}
	log.Printf("Repository URL: %s", repoURL)
	
	revision := branch
	if deployment.CommitSHA != "" {
		revision = deployment.CommitSHA
		log.Printf("Using specific commit SHA: %s", revision)
	}
	
	log.Println("Preparing kpack Image configuration")
	
	spec := map[string]interface{}{
		"tag":                image,
		"serviceAccountName": "default",
		"builder": map[string]interface{}{
			"name": "public-builder",
			"kind": "ClusterBuilder",
		},
		"source": map[string]interface{}{
			"git": map[string]interface{}{
				"url":      repoURL,
				"revision": revision,
			},
		},
	}
	
	// Add environment variables as build-time env
	if len(service.EnvVars) > 0 {
		var env []map[string]interface{}
		for key, value := range service.EnvVars {
			env = append(env, map[string]interface{}{
				"name":  key,
				"value": value,
			})
		}
		spec["build"] = map[string]interface{}{
			"env": env,
		}
	}
	
	kpackImage := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kpack.io/v1alpha2",
			"kind":       "Image",
			"metadata": map[string]interface{}{
				"name":      imageName,
				"namespace": GetJobNamespace(),
				"labels": map[string]interface{}{
					"app":           "pendeploy",
					"service-id":    service.ID,
					"deployment-id": deployment.ID,
					"builder":       "kpack",
					"job-name":      imageName,
				},
			},
			"spec": spec,
		},
	}
	
	log.Println("kpack Image resource created successfully")
	return kpackImage, nil
}