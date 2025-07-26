// @title           Task Manager API
// @version         1.0
// @description     A simple REST API built with Go
// @host            localhost:8080
// @BasePath        /

package main

import (
	"fmt"
	"log"
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"task-api/db"
	_ "task-api/docs"
	"task-api/handlers"
	"task-api/middlewares"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Task Manager API")
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handlers.GetTasks(w, r)
	} else if r.Method == http.MethodPost {
		middlewares.RequireAuth(handlers.CreateTask)(w, r)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	db.Connect()
	defer db.Pool.Close()

	r := mux.NewRouter()

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/", homeHandler)

	// auth handlers
	r.HandleFunc("/login", handlers.Login).Methods("POST")

	// tasks handlers
	r.HandleFunc("/tasks", taskHandler)
	r.HandleFunc("/tasks/{id}", handlers.GetTaskByID).Methods("GET")
	r.HandleFunc("/tasks/{id}", middlewares.RequireAuth(handlers.UpdateTask)).Methods("PUT")
	r.HandleFunc("/tasks/{id}", middlewares.RequireAuth(handlers.DeleteTask)).Methods("DELETE")

	// File upload handler
	r.HandleFunc("/upload", middlewares.RequireAuth(handlers.UploadImage)).Methods("POST")

	// CORS config
	headersOk := gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := gorillaHandlers.AllowedOrigins([]string{"*"})
	methodsOk := gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	fmt.Println("Server starting at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8000", gorillaHandlers.CORS(originsOk, headersOk, methodsOk)(r)))

}
