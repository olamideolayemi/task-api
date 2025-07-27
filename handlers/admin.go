package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"task-api/db"
	"task-api/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(), "SELECT id, email, role FROM users")
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Role); err != nil {
			continue
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetAllTasksWithUsers(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT t.id, t.title, t.details, t.done, t.image_url, u.id, u.email
	FROM tasks t
	JOIN users u ON t.user_id = u.id
	ORDER BY t.id DESC
	`

	rows, err := db.Pool.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TaskWithUser struct {
		ID       uuid.UUID `json:"id"`
		Title    string    `json:"title"`
		Details  string    `json:"details"`
		Done     bool      `json:"done"`
		ImageURL string    `json:"image_url"`
		UserID   uuid.UUID `json:"user_id"`
		Email    string    `json:"email"`
	}

	var result []TaskWithUser
	for rows.Next() {
		var t TaskWithUser
		err := rows.Scan(&t.ID, &t.Title, &t.Details, &t.Done, &t.ImageURL, &t.UserID, &t.Email)
		if err != nil {
			continue
		}
		result = append(result, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Role != "admin" && body.Role != "user") {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(), "UPDATE users SET role=$1 WHERE id=$2", body.Role, id)
	if err != nil {
		http.Error(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// First delete user tasks (if needed)
	_, err = db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE user_id=$1", id)
	if err != nil {
		http.Error(w, "Failed to delete user tasks", http.StatusInternalServerError)
		return
	}

	// Then delete user
	_, err = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetAdminStats(w http.ResponseWriter, r *http.Request) {
	var users, admins, tasks int

	_ = db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&users)
	_ = db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE role='admin'").Scan(&admins)
	_ = db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&tasks)

	stats := map[string]int{
		"total_users": users,
		"admins":      admins,
		"total_tasks": tasks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
