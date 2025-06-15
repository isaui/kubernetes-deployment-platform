package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isabu/pendeploy-handal/models"
	"github.com/isabu/pendeploy-handal/services"
	"github.com/isabu/pendeploy-handal/utils"
)

// GetDeployments menampilkan semua deployment
func GetDeployments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Daftar deployment akan ditampilkan di sini",
	})
}

// GetDeployment menampilkan deployment berdasarkan ID
func GetDeployment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Detail deployment akan ditampilkan di sini",
	})
}

// Services untuk deployment
var (
	deploymentService *services.DeploymentService
	kubernetesService *services.KubernetesService
)

// init menginisialisasi services
func init() {
	deploymentService = services.NewDeploymentService()
	kubernetesService = services.NewKubernetesService()
}

// CreateDeployment membuat deployment baru
func CreateDeployment(c *gin.Context) {
	var req models.DeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ekstrak nama repository dari URL
	repoName := utils.ExtractRepoName(req.RepoURL)

	// Segera kembalikan respons ke client
	responseData := map[string]interface{}{
		"repository": repoName,
		"repoUrl":    req.RepoURL,
		"branch":     req.Branch,
		"commitId":   req.CommitID,
		"message":    "Deployment process started. You can check the logs for status updates.",
		"status":     "in_progress",
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status": "success",
		"data":   responseData,
	})

	// Jalankan semua proses deployment di background
	go func(deploymentReq models.DeploymentRequest) {
		// Log informasi deployment
		fmt.Println("================ DEPLOYMENT REQUEST =================")
		fmt.Printf("Repository: %s\n", repoName)
		fmt.Printf("Repository URL: %s\n", deploymentReq.RepoURL)
		fmt.Printf("Branch: %s\n", deploymentReq.Branch)
		fmt.Printf("Commit ID: %s\n", deploymentReq.CommitID)
		fmt.Println("Git Clone Command: git clone -b " + deploymentReq.Branch + " " + deploymentReq.RepoURL)
		fmt.Println("======================================================")
		
		// Step 1: Clone repository dan timpa data sebelumnya jika ada
		cloneResult, err := deploymentService.CloneRepository(deploymentReq.RepoURL, deploymentReq.Branch)
		if err != nil {
			fmt.Printf("Error cloning repository: %v\n", err)
			return
		}
		
		// Log hasil clone
		fmt.Println("Clone result:", cloneResult)
		
		// Step 2: Build container image dari repository menggunakan nerdctl
		buildResult, err := deploymentService.BuildImage(deploymentReq.RepoURL, deploymentReq.CommitID, deploymentReq.Branch)
		if err != nil {
			fmt.Printf("Error building image: %v\n", err)
			return
		}

		// Log hasil build image
		fmt.Println("================ IMAGE BUILD RESULT =================")
		fmt.Printf("Image name: %s\n", buildResult.ImageName)
		fmt.Printf("Image tag: %s\n", buildResult.ImageTag)
		fmt.Printf("Container ID: %s\n", buildResult.ContainerID)
		fmt.Printf("Build output:\n%s\n", buildResult.Output)
		fmt.Println("=====================================================")
		
		// Step 3: Generate dan jalankan Deploy Script untuk Kubernetes
		repoDir := deploymentService.GetRepoDirectory(deploymentReq.RepoURL)
		scriptPath, err := kubernetesService.GenerateDeployScript(repoDir, buildResult)
		if err != nil {
			fmt.Printf("Warning: Could not generate Kubernetes deploy script: %v\n", err)
		} else {
			// Log hasil generate deploy script
			fmt.Println("================ KUBERNETES DEPLOY SCRIPT =================")
			fmt.Printf("Script path: %s\n", scriptPath)
			fmt.Println("=========================================================")
			
			// Jalankan script deployment Kubernetes
			fmt.Println("Running Kubernetes deployment script...")
			output, deployErr := kubernetesService.RunDeploymentScript(scriptPath)
			if deployErr != nil {
				fmt.Printf("Warning: Failed to run Kubernetes deployment: %v\n", deployErr)
				fmt.Printf("Output: %s\n", output)
			} else {
				fmt.Println("================ KUBERNETES DEPLOYMENT RESULT =================")
				fmt.Printf("Output: %s\n", output)
				fmt.Println("===========================================================")
				fmt.Printf("Kubernetes deployment for %s:%s completed successfully\n", 
					buildResult.ImageName, buildResult.ImageTag)
			}
			// TODO: Di sini bisa ditambahkan kode untuk update status di database
			// atau panggil webhook untuk notifikasi deployment selesai
		}
	}(req)
}

// UpdateDeployment mengupdate deployment
func UpdateDeployment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment akan diupdate di sini",
	})
}

// DeleteDeployment menghapus deployment
func DeleteDeployment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment akan dihapus di sini",
	})
}

// GitHubWebhook menerima webhook dari GitHub
func GitHubWebhook(c *gin.Context) {
	type GitHubCommit struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}

	type GitHubWebhookRequest struct {
		Ref        string `json:"ref"`          // Format: refs/heads/main
		Repository struct {
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			HtmlURL  string `json:"html_url"`
		} `json:"repository"`
		HeadCommit GitHubCommit `json:"head_commit"`
		Commits    []GitHubCommit `json:"commits"`
	}

	var webhookData GitHubWebhookRequest

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract branch name from ref (refs/heads/main -> main)
	branchName := ""
	if len(webhookData.Ref) > 11 { // "refs/heads/" is 11 characters
		branchName = webhookData.Ref[11:]
	}

	// Log data ke console
	fmt.Println("================== GITHUB WEBHOOK RECEIVED ==================")
	fmt.Printf("Repository: %s\n", webhookData.Repository.FullName)
	fmt.Printf("Branch: %s\n", branchName)
	fmt.Printf("Latest Commit: %s\n", webhookData.HeadCommit.ID)
	fmt.Printf("Commit Message: %s\n", webhookData.HeadCommit.Message)
	fmt.Printf("Total Commits: %d\n", len(webhookData.Commits))
	fmt.Println("===========================================================")

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"repository": webhookData.Repository.FullName,
			"branch":     branchName,
			"commit":     webhookData.HeadCommit.ID,
			"message":    webhookData.HeadCommit.Message,
		},
	})
}

// UpdateDeploymentStatus mengupdate status deployment
func UpdateDeploymentStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Status deployment akan diupdate di sini",
	})
}
