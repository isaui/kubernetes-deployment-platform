package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/services"
	"github.com/pendeploy-simple/utils"
)

// CreateDeployment handles the main deployment endpoint
func CreateDeployment(c *gin.Context) {
	var req models.DeploymentRequest
	
	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid request body",
			"error": err.Error(),
		})
		return
	}
	
	// Validate IMAGE_REGISTRY is present
	if err := utils.ValidateImageRegistry(req.Env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "IMAGE_REGISTRY is required in env",
			"error": err.Error(),
		})
		return
	}
	
	// Get repository name for logging
	repoName := utils.ExtractRepoName(req.GithubUrl)
	imageRegistry := req.Env["IMAGE_REGISTRY"]
	
	log.Printf("🚀 Starting deployment for: %s -> %s", repoName, imageRegistry)
	
	// Return immediate response and process in background
	c.JSON(http.StatusAccepted, models.DeploymentResponse{
		Status:    "accepted",
		ImageName: imageRegistry,
		Message:   "Deployment started, processing in background...",
	})
	
	// Process deployment in background goroutine
	go processDeployment(req, repoName)
}

// processDeployment handles the actual deployment process in background
func processDeployment(req models.DeploymentRequest, repoName string) {
	startTime := time.Now()
	imageRegistry := req.Env["IMAGE_REGISTRY"]
	
	log.Printf("================ DEPLOYMENT STARTED ================")
	log.Printf("Repository: %s", req.GithubUrl)
	log.Printf("Image: %s", imageRegistry)
	log.Printf("Environment Variables: %d", len(req.Env))
	log.Printf("====================================================")
	
	// Initialize services
	gitService := services.NewGitService()
	buildService := services.NewBuildService()
	kubernetesService := services.NewKubernetesService()
	
	var cloneDir string
	
	// Cleanup function
	defer func() {
		if cloneDir != "" {
			log.Printf("🧹 Cleaning up temporary directory: %s", cloneDir)
			utils.CleanupDir(cloneDir)
		}
		
		duration := time.Since(startTime)
		log.Printf("⏱️ Total deployment time: %v", duration)
		log.Printf("================ DEPLOYMENT FINISHED ================")
	}()
	
	// Step 1: Clone repository
	log.Printf("🔄 Step 1: Cloning repository...")
	gitResult, err := gitService.CloneRepository(req.GithubUrl)
	if err != nil {
		log.Printf("❌ Git clone failed: %v", err)
		return
	}
	cloneDir = gitResult.CloneDir
	log.Printf("✅ Git clone successful: %s", cloneDir)
	
	// Step 2: Build and push Docker image
	log.Printf("🔨 Step 2: Building and pushing image...")
	buildResult, err := buildService.BuildAndPushImage(cloneDir, req.Env)
	if err != nil {
		log.Printf("❌ Build failed: %v", err)
		log.Printf("Build output:\n%s", buildResult.Output)
		return
	}
	log.Printf("✅ Build and push successful")
	
	// Step 3: Process and apply Kubernetes manifests
	log.Printf("🎯 Step 3: Processing and applying Kubernetes manifests...")
	k8sResult, err := kubernetesService.ProcessAndApplyManifests(cloneDir, req.Env)
	if err != nil {
		log.Printf("❌ Kubernetes deployment failed: %v", err)
		log.Printf("Kubectl output:\n%s", k8sResult.Output)
		return
	}
	log.Printf("✅ Kubernetes deployment successful")
	log.Printf("Kubectl output:\n%s", k8sResult.Output)
	
	// Step 4: Attempt to get service URL
	appName := utils.SanitizeAppName(repoName)
	log.Printf("🔍 Step 4: Checking service URL for app: %s", appName)
	
	// Wait a bit for service to be ready
	time.Sleep(5 * time.Second)
	
	serviceURL, err := kubernetesService.GetServiceURL(appName)
	if err != nil {
		log.Printf("⚠️ Could not get service URL: %v", err)
		serviceURL = "Service URL not available yet"
	} else {
		log.Printf("🌐 Service URL: %s", serviceURL)
	}
	
	// Step 5: Final verification
	log.Printf("🔍 Step 5: Verifying deployment...")
	isReady, err := kubernetesService.VerifyDeployment(appName)
	if err != nil {
		log.Printf("⚠️ Could not verify deployment: %v", err)
	} else if isReady {
		log.Printf("✅ Deployment is ready and running")
	} else {
		log.Printf("⏳ Deployment is still starting up...")
	}
	
	// Log final summary
	log.Printf("================== DEPLOYMENT SUMMARY ==================")
	log.Printf("Repository: %s", req.GithubUrl)
	log.Printf("Image: %s", imageRegistry)
	log.Printf("App Name: %s", appName)
	log.Printf("Service URL: %s", serviceURL)
	log.Printf("Status: %s", func() string {
		if isReady {
			return "READY"
		}
		return "STARTING"
	}())
	log.Printf("Duration: %v", time.Since(startTime))
	log.Printf("========================================================")
}