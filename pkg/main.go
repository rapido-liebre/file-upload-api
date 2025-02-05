package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	_ "file-upload-api/docs"
)

// Directory for storing uploaded files
const uploadDir = "./uploads"

// @title File Upload API
// @version 1.0
// @description API for uploading and listing files with metadata
// @host localhost:4000
// @BasePath /
// @schemes http

func main() {
	// Ensure the upload directory exists
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key"},
		AllowCredentials: true,
	}))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// File upload endpoint
	r.POST("/upload", AuthMiddleware(), UploadFile)

	// File listing endpoint
	r.GET("/files", AuthMiddleware(), ListFiles)

	// Start the server on port 4000
	r.Run(":4000")
}

// Hardcoded API key for simple authentication
const validAPIKey = "my-secret-api-key"

// AuthMiddleware to check API Key in request headers
// @Summary Authenticate API Key
// @Description Middleware that validates API Key
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /upload [post]
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != validAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid API Key"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ListFiles handles listing uploaded files
// @Summary List uploaded files
// @Description Retrieves a list of uploaded files and their metadata
// @Tags Files
// @Produce json
// @Success 200 {array} map[string]string
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /files [get]
func ListFiles(c *gin.Context) {
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}

	var fileList []gin.H

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".metadata" {
			continue
		}

		metadataPath := filepath.Join(uploadDir, file.Name()+".metadata")
		var metadata map[string]string

		if metaFile, err := os.ReadFile(metadataPath); err == nil {
			_ = json.Unmarshal(metaFile, &metadata)
		}

		fileList = append(fileList, gin.H{
			"file_name":   file.Name(),
			"file_path":   "/uploads/" + file.Name(),
			"title":       metadata["title"],
			"description": metadata["description"],
		})
	}

	c.JSON(http.StatusOK, fileList)
}

// UploadFile handles file uploads
// @Summary Upload a file
// @Description Uploads a file and stores metadata
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param title formData string false "File title"
// @Param description formData string false "File description"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /upload [post]
func UploadFile(c *gin.Context) {
	log.Println("Processing file upload request...")

	title := c.PostForm("title")
	description := c.PostForm("description")

	header, err := c.FormFile("file")
	if err != nil {
		log.Println("Error retrieving file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file"})
		return
	}

	originalFileName := header.Filename
	filePath := filepath.Join(uploadDir, originalFileName)
	log.Printf("Saving file as: %s\n", filePath)

	if err := c.SaveUploadedFile(header, filePath); err != nil {
		log.Println("Error saving file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	log.Println("File successfully saved.")

	metadata := map[string]string{
		"title":       title,
		"description": description,
		"filename":    originalFileName,
	}

	metadataFilePath := filePath + ".metadata"
	log.Printf("Saving metadata file: %s\n", metadataFilePath)

	metadataFile, err := os.Create(metadataFilePath)
	if err != nil {
		log.Println("Error creating metadata file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}
	defer metadataFile.Close()

	encoder := json.NewEncoder(metadataFile)
	if err := encoder.Encode(metadata); err != nil {
		log.Println("Error writing metadata file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write metadata"})
		return
	}
	log.Println("Metadata successfully saved.")

	c.JSON(http.StatusOK, gin.H{
		"message":       "File uploaded successfully",
		"file_name":     originalFileName,
		"file_path":     filePath,
		"metadata_path": metadataFilePath,
		"title":         title,
		"description":   description,
	})
}
