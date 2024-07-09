package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

type Todo struct {
	ID     int    `json:"id"`
	Task   string `json:"task"`
	UserID string `json:"user_id"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal(err)
	}

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT
	);`

	createTodosTable := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT,
		user_id TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	if _, err := db.Exec(createUsersTable); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(createTodosTable); err != nil {
		log.Fatal(err)
	}

	// Insert users if not already present
	users := []User{
		{ID: "shelley", Name: "Shelley"},
		{ID: "keagan", Name: "Keagan"},
		{ID: "dane", Name: "Dane"},
		{ID: "paul", Name: "Paul"},
	}
	for _, user := range users {
		if _, err := db.Exec("INSERT OR IGNORE INTO users (id, name) VALUES (?, ?)", user.ID, user.Name); err != nil {
			log.Fatal(err)
		}
	}
}

func GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := getAllTodos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(todos)
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := addTodo(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := removeTodoByID(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getAllTodos() ([]Todo, error) {
	rows, err := db.Query("SELECT id, task, user_id FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Task, &todo.UserID); err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func addTodo(todo *Todo) error {
	result, err := db.Exec("INSERT INTO todos (task, user_id) VALUES (?, ?)", todo.Task, todo.UserID)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	todo.ID = int(id)
	return nil
}

func removeTodoByID(id int) error {
	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	return err
}

func main() {
	initDB()

	router := mux.NewRouter()
	router.HandleFunc("/todos", GetTodos).Methods("GET")
	router.HandleFunc("/todos", CreateTodo).Methods("POST")
	router.HandleFunc("/todos/{id}", DeleteTodo).Methods("DELETE")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	})

	handler := c.Handler(router)

	log.Println("Server started on :9090")
	log.Fatal(http.ListenAndServe(":9090", handler))
}
