package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID        int
	Completed bool
	Content   string
}

var tmpl *template.Template
var db *sql.DB

func main() {
	var err error
	tmpl, err = template.ParseFiles("templates.html")

	if err != nil {
		panic(err)
	}

	db, err = sql.Open("sqlite3", "file:sqlite.db")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		completed BOOLEAN NOT NULL,
		content TEXT NOT NULL
	)`)

	if err != nil {
		panic(err)
	}

	statement.Exec()

	port := os.Getenv("PORT")

	if port == "" {
		port = "3333"
	}

	r := chi.NewRouter()
	addr := ":" + port

	r.Use(middleware.Logger)
	r.Get("/", RootRoute)
	r.Get("/todos", TodosRoute)
	r.Post("/todos", TodosRoute)
	r.Patch("/todos/{id}", TodoRoute)
	fmt.Printf("Server is running at http://localhost%v\n", addr)
	http.ListenAndServe(addr, r)
}

func RootRoute(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "root", nil)
}

func TodosRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		content := r.FormValue("content")
		statement, err := db.Prepare("INSERT INTO todos (completed, content) VALUES (?, ?)")

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		statement.Exec(false, content)
	}

	rows, err := db.Query("SELECT * FROM todos")

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo

		rows.Scan(&todo.ID, &todo.Completed, &todo.Content)

		todos = append(todos, todo)
	}

	tmpl.ExecuteTemplate(w, "todos-list", todos)
}

func TodoRoute(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if r.Method == "PATCH" {
		r.ParseForm()

		completed := r.FormValue("completed") == "on"
		statement, err := db.Prepare("UPDATE todos SET completed = ? WHERE id = ?")

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		statement.Exec(completed, id)
	}

	row := db.QueryRow("SELECT * FROM todos WHERE id = ?", id)
	var todo Todo

	row.Scan(&todo.ID, &todo.Completed, &todo.Content)
	tmpl.ExecuteTemplate(w, "todo-item", todo)
}
