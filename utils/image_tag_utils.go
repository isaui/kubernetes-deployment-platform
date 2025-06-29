package utils

import (
	"fmt"
	"log"

	"github.com/pendeploy-simple/models"
)

func GenerateImage(registryURL string, service models.Service, deployment models.Deployment) string {
    log.Println("Generating image tag for service: " + service.Name + ", deployment: " + deployment.ID)
	log.Printf("Image tag: %s", fmt.Sprintf("%s/%s:%s", CleanRegistryURL(registryURL), service.ID, deployment.ID))
	return fmt.Sprintf("%s/%s:%s", CleanRegistryURL(registryURL), service.ID, deployment.ID)
}
