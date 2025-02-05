
---

## **📄 `backend/README.md`**
```markdown
# Backend - File Upload API

This is the **backend** for the file upload system, built with **Go + Gin**.  
It provides an API for **file uploads, metadata storage, and listing files**.

## 🚀 Features
- Upload files with metadata (title & description)
- Store metadata in separate `.metadata` files
- List all uploaded files
- API protected with an API Key
- OpenAPI (Swagger) documentation

## 📦 Requirements
- **Go** (v1.23+)
- **Docker** (for containerized deployment)

---

## 🛠️ Installation & Running Locally

### **1️⃣ Install Dependencies**
```sh
cd backend
go mod tidy

