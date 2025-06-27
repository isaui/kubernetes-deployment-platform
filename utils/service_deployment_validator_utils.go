package utils

import "github.com/pendeploy-simple/models"

func ValidateServiceDeployment(service models.Service, apiKey string) (bool, error) {
	if(service.APIKey != apiKey) {
		return false, nil
	}
	return true, nil
}