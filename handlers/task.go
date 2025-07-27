package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"task-api/db"
	"task-api/middlewares"
	"task-api/models"
	"task-api/utils"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
)

// var tasks = []models.Task{}
// var nextID = 1

// Get all tasks
func GetTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(), "SELECT id, title, details, done, image_url FROM tasks")
	if err != nil {
		log.Printf("GetTask error: %v", err)
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Details, &task.Done, &task.ImageURL)
		if err != nil {
			http.Error(w, "Failed to parse task", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// CreateTask godoc
// @Summary      Create a new task with optional image
// @Description  Adds a task and uploads image to Cloudinary
// @Tags         tasks
// @Accept       mpfd
// @Produce      json
// @Param        title formData string true "Task title"
// @Param        details formData string false "Task details"
// @Param        done formData boolean false "Is task done?"
// @Param        image formData file false "Image file to upload"
// @Success      201 {object} models.Task
// @Failure      400 {string} string "Bad request"
// @Failure      500 {string} string "Internal error"
// @Router       /tasks [post]
func CreateTask(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	details := r.FormValue("details")
	doneStr := r.FormValue("done")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	done := false
	if strings.ToLower(doneStr) == "true" {
		done = true
	}

	var imageURL string

	// Optional image upload
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		cld, err := cloudinary.NewFromParams(
			os.Getenv("CLOUDINARY_CLOUD_NAME"),
			os.Getenv("CLOUDINARY_API_KEY"),
			os.Getenv("CLOUDINARY_API_SECRET"),
		)
		if err != nil {
			http.Error(w, "Cloudinary config failed", http.StatusInternalServerError)
			return
		}

		uploadResp, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{})
		if err != nil {
			http.Error(w, "Failed to upload image", http.StatusInternalServerError)
			return
		}

		imageURL = uploadResp.SecureURL
	}

	userID := middlewares.GetUserID(r)
	var task models.Task
	err = db.Pool.QueryRow(
		context.Background(),
		`INSERT INTO tasks (title, details, done, image_url, user_id)
	 VALUES ($1, $2, $3, $4, $5)
	 RETURNING id, title, details, done, image_url, user_id`,
		title, details, done, imageURL, userID,
	).Scan(&task.ID, &task.Title, &task.Details, &task.Done, &task.ImageURL, &task.UserID)

	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	var userEmail string
	err = db.Pool.QueryRow(context.Background(),
		"SELECT email FROM users WHERE id=$1", task.UserID).Scan(&userEmail)
	if err == nil {
		go utils.SendEmail(userEmail, "New Task Created", fmt.Sprintf("Hi, your task '%s' has been created!", task.Title))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// Get Tak by ID
func GetTaskByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	err = db.Pool.QueryRow(
		context.Background(),
		"SELECT id, title, details, done, image_url FROM tasks WHERE id=$1", id,
	).Scan(&task.ID, &task.Title, &task.Details, &task.Done, &task.ImageURL)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask godoc
// @Summary      Update a task (including image)
// @Description  Updates task fields and optionally replaces the image
// @Tags         tasks
// @Accept       mpfd
// @Produce      json
// @Param        id path int true "Task ID"
// @Param        title formData string true "Task title"
// @Param        details formData string false "Task details"
// @Param        done formData boolean false "Done status"
// @Param        image formData file false "New image file"
// @Success      200 {object} models.Task
// @Failure      400 {string} string "Bad request"
// @Failure      404 {string} string "Task not found"
// @Router       /tasks/{id} [put]
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	details := r.FormValue("details")
	doneStr := r.FormValue("done")
	done := false
	if strings.ToLower(doneStr) == "true" {
		done = true
	}

	imageURL := ""

	// Optional image upload
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		cld, err := cloudinary.NewFromParams(
			os.Getenv("CLOUDINARY_CLOUD_NAME"),
			os.Getenv("CLOUDINARY_API_KEY"),
			os.Getenv("CLOUDINARY_API_SECRET"),
		)
		if err != nil {
			http.Error(w, "Cloudinary config failed", http.StatusInternalServerError)
			return
		}

		uploadResp, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{})
		if err != nil {
			http.Error(w, "Image upload failed", http.StatusInternalServerError)
			return
		}

		imageURL = uploadResp.SecureURL
	}

	var commandTag pgconn.CommandTag
	if imageURL != "" {
		commandTag, err = db.Pool.Exec(context.Background(),
			"UPDATE tasks SET title=$1, details=$2, done=$3, image_url=$4 WHERE id=$5",
			title, details, done, imageURL, id,
		)
	} else {
		commandTag, err = db.Pool.Exec(context.Background(),
			"UPDATE tasks SET title=$1, details=$2, done=$3 WHERE id=$4",
			title, details, done, id,
		)
	}

	if err != nil || commandTag.RowsAffected() == 0 {
		http.Error(w, "Task not found or update failed", http.StatusNotFound)
		return
	}

	// Return updated task
	var updatedTask models.Task
	err = db.Pool.QueryRow(
		context.Background(),
		"SELECT id, title, details, done, image_url FROM tasks WHERE id=$1", id,
	).Scan(&updatedTask.ID, &updatedTask.Title, &updatedTask.Details, &updatedTask.Done, &updatedTask.ImageURL)

	if err != nil {
		http.Error(w, "Failed to fetch updated task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// Delete task
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := middlewares.GetUserID(r)
	role := middlewares.GetUserRole(r)

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if role != "admin" {
		var ownerID int
		err := db.Pool.QueryRow(context.Background(),
			"SELECT user_id FROM tasks WHERE id=$1", id,
		).Scan(&ownerID)

		if err != nil || ownerID != userID {
			http.Error(w, "Not authorized to delete this task", http.StatusForbidden)
			return
		}
	}

	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	commandTag, err := db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE id=$1", id)

	if err != nil || commandTag.RowsAffected() == 0 {
		http.Error(w, "Delete failed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// Get User Tasks
func GetUserTasks(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, _ := strconv.Atoi(params["id"])

	rows, err := db.Pool.Query(context.Background(), "SELECT id, title, details, done, image_url FROM tasks WHERE user_id=$1", userID)
	if err != nil {
		http.Error(w, "Error fetching tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		rows.Scan(&task.ID, &task.Title, &task.Details, &task.Done, &task.ImageURL)
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
