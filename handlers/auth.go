package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"task-api/db"
	"task-api/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var user models.User
	var role string
	err := db.Pool.QueryRow(context.Background(),
		"SELECT id, password, role FROM users WHERE email=$1", req.Email,
	).Scan(&user.ID, &user.Password, &role)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// generate JWT
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"user_id":  user.ID,
		"role":     role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Token error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var user models.User
	_ = json.NewDecoder(r.Body).Decode(&user)

	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(user.Email)
	if user.Name == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "All fields required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	err = db.Pool.QueryRow(
		context.Background(),
		`INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`,
		user.Name, user.Email, string(hashed),
	).Scan(&user.ID)

	if err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	json.NewEncoder(w).Encode(user)
}
