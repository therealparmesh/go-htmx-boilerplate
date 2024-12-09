package main

import (
	"database/sql"
	"encoding/json"
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

type TriggerHeader struct {
	SuccessNotification *string `json:"successNotification,omitempty"`
	ErrorNotification   *string `json:"errorNotification,omitempty"`
}

var tmpl *template.Template
var db *sql.DB

func main() {
	var err error
	tmpl, err = template.ParseFiles("main.gohtml")
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
	addr := "localhost:" + port
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(9))
	if os.Getenv("NODE_ENV") == "development" {
		r.Use(middleware.NoCache)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("build"))))
	r.Get("/", RootRoute)
	r.Get("/todos", TodosRoute)
	r.Post("/todos", TodosRoute)
	r.Patch("/todos/{id}", TodoRoute)
	r.Delete("/todos/{id}", TodoRoute)
	fmt.Printf("Server is running at http://%v\n", addr)
	http.ListenAndServe(addr, r)
}

func RootRoute(w http.ResponseWriter, _ *http.Request) {
	tmpl.ExecuteTemplate(w, "root", nil)
}

func TodosRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		content := r.FormValue("content")
		statement, err := db.Prepare("INSERT INTO todos (completed, content) VALUES (?, ?)")
		if err != nil {
			setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not added!")})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result, err := statement.Exec(false, content)
		if err != nil {
			setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not added!")})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not added!")})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var todo Todo
		row := db.QueryRow("SELECT * FROM todos WHERE id = ?", id)
		err = row.Scan(&todo.ID, &todo.Completed, &todo.Content)
		if err != nil {
			setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not added!")})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		setTriggerHeader(w, TriggerHeader{SuccessNotification: strPtr("Todo added successfully!")})
		tmpl.ExecuteTemplate(w, "todo-item", todo)
		return
	}

	var todos []Todo
	rows, err := db.Query("SELECT * FROM todos")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo

		err := rows.Scan(&todo.ID, &todo.Completed, &todo.Content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		todos = append(todos, todo)
	}

	tmpl.ExecuteTemplate(w, "todos-list", todos)
}

func TodoRoute(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "DELETE" {
		statement, err := db.Prepare("DELETE FROM todos WHERE id = ?")
		if err != nil {
			setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not deleted!")})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		statement.Exec(id)
		setTriggerHeader(w, TriggerHeader{SuccessNotification: strPtr("Todo deleted successfully!")})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	r.ParseForm()

	completed := r.FormValue("completed") == "on"
	statement, err := db.Prepare("UPDATE todos SET completed = ? WHERE id = ?")
	if err != nil {
		setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not updated!")})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	statement.Exec(completed, id)

	var todo Todo
	row := db.QueryRow("SELECT * FROM todos WHERE id = ?", id)
	err = row.Scan(&todo.ID, &todo.Completed, &todo.Content)
	if err != nil {
		setTriggerHeader(w, TriggerHeader{ErrorNotification: strPtr("Todo not updated!")})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setTriggerHeader(w, TriggerHeader{SuccessNotification: strPtr("Todo updated successfully!")})
	tmpl.ExecuteTemplate(w, "todo-item", todo)
}

func setTriggerHeader(w http.ResponseWriter, t TriggerHeader) {
	j, err := json.Marshal(t)
	if err != nil {
		w.Header().Set("hx-trigger", "")
		return
	}

	w.Header().Set("hx-trigger", string(j))
}

func strPtr(s string) *string {
	return &s
}
