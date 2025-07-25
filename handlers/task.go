// CreateTask godoc
// @Summary      Create a new task
// @Description  Adds a task to the database
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body  models.Task  true  "Task to create"
// @Success      200   {object}  models.Task
// @Failure      400   {string}  string  "Bad request"
// @Failure      500   {string}  string  "Internal error"
// @Router       /tasks [post]

package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"task-api/db"
	"task-api/models"

	"github.com/gorilla/mux"
)

// var tasks = []models.Task{}
// var nextID = 1

// Get all tasks
func GetTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(), "SELECT id, title, details, done FROM tasks")
	if err != nil {
		log.Printf("GetTask error: %v", err)
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Details, &task.Done)
		if err != nil {
			http.Error(w, "Failed to parse task", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// Create a new task
func CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	err := db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO tasks (title, details, done) VALUES ($1, $2, $3) RETURNING id",
		task.Title, task.Details, task.Done,
	).Scan(&task.ID)

	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
		"SELECT id, title, details, done FROM tasks WHERE id=$1", id,
	).Scan(&task.ID, &task.Title, &task.Details, &task.Done)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// Update Task
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	var updatedTask models.Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	updatedTask.Title = strings.TrimSpace(updatedTask.Title)
	if updatedTask.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	commandTag, err := db.Pool.Exec(
		context.Background(),
		"UPDATE tasks SET title=$1, details=$2, done=$3 WHERE id=$4",
		updatedTask.Title, updatedTask.Details, updatedTask.Done, id,
	)

	if err != nil || commandTag.RowsAffected() == 0 {
		http.Error(w, "Task not found or update failed", http.StatusNotFound)
		return
	}

	updatedTask.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

// Delete task
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	commandTag, err := db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE id=$1", id)

	if err != nil || commandTag.RowsAffected() == 0 {
		http.Error(w, "Task not found or delete failed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
