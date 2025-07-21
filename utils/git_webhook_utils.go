package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// SendWebhookNotification sends a notification to a webhook URL with deployment status and optional error message
func SendWebhookNotification(webhookUrl string, deploymentID string, status string, errorMessage string) {
	// If no webhook URL is provided, do nothing
	if webhookUrl == "" {
		return
	}
	
	// Sanitize the webhook URL - remove any whitespace or newlines
	webhookUrl = strings.TrimSpace(webhookUrl)
	
	// Safety check for deploymentID
	if deploymentID == "" {
		log.Printf("Warning: Empty deploymentID in webhook notification")
	}
	
	// Clean error message if present to prevent JSON parsing errors
	if errorMessage != "" {
		errorMessage = strings.ReplaceAll(errorMessage, "\n", " ")
	}
	
	// Prepare webhook payload
	payload := map[string]interface{}{
		"deploymentId": deploymentID,
		"status":      status,
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	
	// Add error message if provided
	if errorMessage != "" {
		payload["error"] = errorMessage
	}
	
	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling webhook payload: %v", err)
		return
	}
	
	// Send webhook request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error calling webhook: %v", err)
		return
	}
	defer resp.Body.Close()
	
	log.Printf("Webhook notification sent to %s, status: %s, deployment: %s", 
		webhookUrl, status, deploymentID)
}

// SendErrorWebhook sends an error notification to a webhook URL (no deployment ID)
func SendErrorWebhook(webhookUrl string, errMessage string) {
	// If no webhook URL is provided, do nothing
	if webhookUrl == "" {
		return
	}
	
	// Sanitize the webhook URL - remove any whitespace or newlines
	webhookUrl = strings.TrimSpace(webhookUrl)
	
	// Clean error message - replace newlines with spaces to prevent JSON parsing errors
	errMessage = strings.ReplaceAll(errMessage, "\n", " ")
	
	// Prepare webhook payload
	payload := map[string]interface{}{
		"status":    "failed",
		"error":     errMessage,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling webhook error payload: %v", err)
		return
	}
	
	// Send webhook request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error calling webhook: %v", err)
		return
	}
	defer resp.Body.Close()
	
	log.Printf("Error webhook notification sent to %s", webhookUrl)
}
