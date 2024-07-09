package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "todo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create Todo table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todo (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT
	);`)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/todos", getTodos).Methods("GET")
	r.HandleFunc("/todos", addTodo).Methods("POST")

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	log.Println("Server listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(r)))
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title FROM todo ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	todos := make([]Todo, 0)
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO todo (title) VALUES (?)", todo.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	todo.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}
