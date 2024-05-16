package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

var errorAuth bool = false
var errorLanding bool = false
var errorPost bool = false
var errorCategories bool = false
var errorInside bool = false
var errorAbout bool = false
var errorCreate bool = false
var errorUser bool = false
var errorParameter bool = false

func auth(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, password)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/auth.html"))
	tmpl.Execute(w, errorAuth)
}

func landing(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/landing.html"))
	tmpl.Execute(w, errorLanding)
}

func post(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/post.html"))
	tmpl.Execute(w, errorPost)
	/*input := r.Form.Get("Input")
	  if input == "Text" {
	      fmt.Println("Ya des soucis")
	  }*/
}

func categories(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/categories.html"))
	tmpl.Execute(w, errorCategories)
}

func inside(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/inside.html"))
	tmpl.Execute(w, errorInside)
}

func create(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/newtopic.html"))
	tmpl.Execute(w, errorCreate)
}

func about(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/about.html"))
	tmpl.Execute(w, errorAbout)
}

func user(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/user.html"))
	tmpl.Execute(w, errorUser)
}

func parameter(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/parameter.html"))
	tmpl.Execute(w, errorParameter)
}

func main() {
	SetupDatabase()
	http.Handle("/home", http.HandlerFunc(landing))
	http.Handle("/auth", http.HandlerFunc(auth))
	http.Handle("/post", http.HandlerFunc(post))
	http.Handle("/categories", http.HandlerFunc(categories))
	http.Handle("/inside", http.HandlerFunc(inside))
	http.Handle("/about", http.HandlerFunc(about))
	http.Handle("/create", http.HandlerFunc(create))
	http.Handle("/user", http.HandlerFunc(user))
	http.Handle("/parameter", http.HandlerFunc(parameter))
	Open("http://localhost/home")
	http.ListenAndServe("", nil)
}

func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func SetupDatabase() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		email TEXT UNIQUE,
		password TEXT NOT NULL,
		admin BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database setup completed.")
}
