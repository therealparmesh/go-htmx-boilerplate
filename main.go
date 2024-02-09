package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
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

var db *sql.DB

func executeTemplate(w http.ResponseWriter, data any, files ...string) error {
	tmpl, err := template.ParseFiles(files...)

	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func writeError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func main() {
	var err error
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

	_, err = statement.Exec()

	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	addr := ":3333"

	r.Use(middleware.Logger)
	r.Get("/", RootRoute)
	r.Get("/todos", TodosRoute)
	r.Post("/todos", TodosRoute)
	r.Patch("/todos/{id}", TodoRoute)
	fmt.Printf("Server is running at http://localhost%v\n", addr)

	err = http.ListenAndServe(addr, r)

	if err != nil {
		panic(err)
	}
}

func RootRoute(w http.ResponseWriter, r *http.Request) {
	err := executeTemplate(w, nil, "templates/index.html", "templates/root.html")

	if err != nil {
		writeError(w, http.StatusInternalServerError)
	}
}

func TodosRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()

		if err != nil {
			writeError(w, http.StatusBadRequest)
			return
		}

		content := r.FormValue("content")
		statement, err := db.Prepare("INSERT INTO todos (completed, content) VALUES (?, ?)")

		if err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(false, content)

		if err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}
	}

	rows, err := db.Query("SELECT * FROM todos")

	if err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var todos []Todo

	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.Completed, &todo.Content)

		if err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}

		todos = append(todos, todo)
	}

	err = executeTemplate(w, todos, "templates/todos-list.html", "templates/todo-item.html")

	if err != nil {
		writeError(w, http.StatusInternalServerError)
	}
}

func TodoRoute(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		writeError(w, http.StatusBadRequest)
		return
	}

	if r.Method == "PATCH" {
		err := r.ParseForm()

		if err != nil {
			writeError(w, http.StatusBadRequest)
			return
		}

		completed := r.FormValue("completed") == "on"
		statement, err := db.Prepare("UPDATE todos SET completed = ? WHERE id = ?")

		if err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(completed, id)

		if err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}
	}

	row := db.QueryRow("SELECT * FROM todos WHERE id = ?", id)
	var todo Todo
	err = row.Scan(&todo.ID, &todo.Completed, &todo.Content)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound)
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}

	err = executeTemplate(w, todo, "templates/todos-list-item.html", "templates/todo-item.html")

	if err != nil {
		writeError(w, http.StatusInternalServerError)
	}
}
