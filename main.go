package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

var errorAuth bool
var errorLanding bool
var errorPost bool
var errorCategories bool
var errorInside bool
var errorAbout bool
var errorCreate bool
var errorUser bool
var errorParameter bool

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

	// Affichage du formulaire d'authentification
	tmpl := template.Must(template.ParseFiles("src/templates/auth.html"))
	if err := tmpl.Execute(w, errorAuth); err != nil {
		log.Println("Erreur lors de l'exécution du template auth:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func landingHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/landing.html"))
	if err := tmpl.Execute(w, errorLanding); err != nil {
		log.Println("Erreur lors de l'exécution du template landing:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/post.html"))
	if err := tmpl.Execute(w, errorPost); err != nil {
		log.Println("Erreur lors de l'exécution du template post:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func categories(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/categories.html"))
	if err := tmpl.Execute(w, errorCategories); err != nil {
		log.Println("Erreur lors de l'exécution du template categories:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func inside(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/inside.html"))
	if err := tmpl.Execute(w, errorInside); err != nil {
		log.Println("Erreur lors de l'exécution du template inside:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func create(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/newtopic.html"))
	if err := tmpl.Execute(w, errorCreate); err != nil {
		log.Println("Erreur lors de l'exécution du template create:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func about(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/about.html"))
	if err := tmpl.Execute(w, errorAbout); err != nil {
		log.Println("Erreur lors de l'exécution du template about:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func user(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/user.html"))
	if err := tmpl.Execute(w, errorUser); err != nil {
		log.Println("Erreur lors de l'exécution du template user:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func parameter(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/parameter.html"))
	if err := tmpl.Execute(w, errorParameter); err != nil {
		log.Println("Erreur lors de l'exécution du template parameter:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[1:] // Remove the leading '/'
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, filePath)
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

func main() {
	SetupDatabase()

	http.Handle("/src/img/", http.StripPrefix("/src/img/", http.FileServer(http.Dir("./src/img"))))
	http.Handle("/src/css/", http.StripPrefix("/src/css/", http.FileServer(http.Dir("./src/css"))))
	http.HandleFunc("/home", landingHandler)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/post", post)
	http.HandleFunc("/categories", categories)
	http.HandleFunc("/inside", inside)
	http.HandleFunc("/about", about)
	http.HandleFunc("/create", create)
	http.HandleFunc("/user", user)
	http.HandleFunc("/parameter", parameter)

	log.Println("Server starting on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Erreur lors du démarrage du serveur HTTP:", err)
		}
	}()

	if err := Open("http://localhost:8080/home"); err != nil {
		log.Println("Erreur lors de l'ouverture du navigateur:", err)
	}

	select {}
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
