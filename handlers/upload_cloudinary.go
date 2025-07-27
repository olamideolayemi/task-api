package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	// "yourproject/utils" // Adjust this import to your project structure

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(w http.ResponseWriter, r *http.Request) {
	// Parse file
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image not provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create Cloudinary instance
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		http.Error(w, "Cloudinary init error", http.StatusInternalServerError)
		return
	}

	// Upload to Cloudinary
	uploadResp, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{})
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the secure URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Image uploaded", "url": "%s"}`, uploadResp.SecureURL)
}
