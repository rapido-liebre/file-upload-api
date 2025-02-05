package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Setup a test router
func setupTestRouter() *gin.Engine {
	r := gin.Default()
	r.Use(AuthMiddleware())
	r.POST("/upload", UploadFile)
	r.GET("/files", ListFiles)
	return r
}

// Test AuthMiddleware - Valid & Invalid API Key
func TestAuthMiddleware(t *testing.T) {
	router := gin.Default()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
	})

	tests := []struct {
		name       string
		apiKey     string
		wantStatus int
	}{
		{"Valid API Key", validAPIKey, http.StatusOK},
		{"Invalid API Key", "wrong-key", http.StatusUnauthorized},
		{"No API Key", "", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("X-API-Key", tt.apiKey)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// Test File Upload
func TestUploadFile(t *testing.T) {
	// Create a temporary test upload directory
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	defer os.RemoveAll(uploadDir) // Cleanup after test

	// Setup test server
	router := setupTestRouter()

	// Create a test file
	testFileName := "testfile.txt"
	testFilePath := filepath.Join(uploadDir, testFileName)
	testMetadataPath := testFilePath + ".metadata" // Ensure this is used

	_ = os.WriteFile(testFilePath, []byte("test content"), os.ModePerm)

	// Prepare multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add metadata fields
	_ = writer.WriteField("title", "Test File")
	_ = writer.WriteField("description", "A test file upload.")

	// Add file
	part, _ := writer.CreateFormFile("file", testFileName)
	testFile, _ := os.Open(testFilePath)
	_, _ = io.Copy(part, testFile)
	testFile.Close()
	writer.Close()

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", validAPIKey) // Valid API Key

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &response)

	// Validate that the file and metadata were saved
	_, fileErr := os.Stat(response["file_path"])
	assert.NoError(t, fileErr, "Uploaded file should exist")

	_, metaErr := os.Stat(testMetadataPath) // Now using testMetadataPath
	assert.NoError(t, metaErr, "Metadata file should exist")

	// Validate metadata content
	metaFile, _ := os.ReadFile(testMetadataPath) // Now using testMetadataPath
	var metaData map[string]string
	_ = json.Unmarshal(metaFile, &metaData)

	assert.Equal(t, "Test File", metaData["title"])
	assert.Equal(t, "A test file upload.", metaData["description"])
	assert.Equal(t, testFileName, metaData["filename"])
}

// Test Listing Files with API Key
func TestListFiles(t *testing.T) {
	// Create a temporary test upload directory
	_ = os.MkdirAll(uploadDir, os.ModePerm)
	defer os.RemoveAll(uploadDir) // Cleanup after test

	// Create a dummy test file
	testFileName := "listfile.txt"
	testFilePath := filepath.Join(uploadDir, testFileName)
	testMetadataPath := testFilePath + ".metadata"

	_ = os.WriteFile(testFilePath, []byte("dummy data"), os.ModePerm)

	// Create dummy metadata
	metaData := map[string]string{
		"title":       "Test File",
		"description": "Test Description",
		"filename":    testFileName,
	}
	metaFile, _ := os.Create(testMetadataPath)
	defer metaFile.Close()
	json.NewEncoder(metaFile).Encode(metaData)

	// Setup test server
	router := setupTestRouter()

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	req.Header.Set("X-API-Key", validAPIKey) // ✅ Add API Key

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response status
	assert.Equal(t, http.StatusOK, w.Code, "ListFiles should return 200 OK")

	// Parse response
	var files []map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &files)

	// Ensure at least one file exists in the response
	assert.Greater(t, len(files), 0, "At least one file should be listed")

	// Validate first file metadata
	assert.Equal(t, "Test File", files[0]["title"])
	assert.Equal(t, "Test Description", files[0]["description"])
	assert.Equal(t, testFileName, files[0]["file_name"])
}
