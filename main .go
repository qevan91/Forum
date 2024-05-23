package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//var isCo bool = false

var errorAuth bool = false
var errorAbout bool = false
var errorCreate bool = false
var errorUser bool = false
var errorParameter bool = false

type Session struct {
	Username string
	Expiry   time.Time
}

type Post struct {
	ID         int
	UserID     int
	CategoryID int
	Content    string
	CreatedAt  time.Time
}

var sessionStore = struct {
	sync.RWMutex
	sessions map[string]Session
}{
	sessions: make(map[string]Session),
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func createSession(w http.ResponseWriter, username string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(24 * time.Hour)

	sessionStore.Lock()
	sessionStore.sessions[sessionID] = Session{
		Username: username,
		Expiry:   expiry,
	}
	sessionStore.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionID,
		Expires: expiry,
	})

	return sessionID, nil
}

func validateSession(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	sessionStore.RLock()
	session, exists := sessionStore.sessions[cookie.Value]
	sessionStore.RUnlock()

	if !exists || session.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("session invalid or expired")
	}

	return &session, nil
}

func deleteSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return
	}

	sessionStore.Lock()
	delete(sessionStore.sessions, cookie.Value)
	sessionStore.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})
}

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

		if username != "" && email != "" && password != "" {
			_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, password)
			if err != nil {
				log.Fatal(err)
			}

			_, err = createSession(w, username)
			if err != nil {
				http.Error(w, "Impossible de créer la session", http.StatusInternalServerError)
				return
			}
		}

		if username != "" && email == "" && password != "" {
			var storedPassword string
			err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Erreur de serveur", http.StatusInternalServerError)
				return
			}

			if password != storedPassword {
				http.Error(w, "Mot de passe incorrect", http.StatusUnauthorized)
				return
			}

			_, err = createSession(w, username)
			if err != nil {
				http.Error(w, "Impossible de créer la session", http.StatusInternalServerError)
				return
			}
		}

		if username == "" && email == "" && password == "" {
			deleteSession(w, r)
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/auth.html"))
	tmpl.Execute(w, errorAuth)
}

type LandingData struct {
	Username string
	Error    string
}

func landing(w http.ResponseWriter, r *http.Request) {
	session, err := validateSession(r)
	data := LandingData{}

	if err != nil {
		data.Error = "Vous devez vous connecter"
	} else {
		data.Username = session.Username
	}

	tmpl := template.Must(template.ParseFiles("src/templates/landing.html"))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	categories, err := getCategories()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting categories", http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories []string
	}{
		Categories: categories,
	}

	if r.Method == "POST" {
		category := r.FormValue("categories")
		message := r.FormValue("message")

		categoryID, err := getCategoryIDByName(category)
		if err != nil {
			http.Error(w, "Can't access have the name", http.StatusInternalServerError)
			return
		}

		session, err := validateSession(r)
		if err != nil {
			http.Error(w, "You must be logged in to post", http.StatusUnauthorized)
			return
		}

		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Println(err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var userName string
		err = db.QueryRow("SELECT username FROM users WHERE username = ?", session.Username).Scan(&userName)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO posts (user_id, category_id, content) VALUES (?, ?, ?)", userName, categoryID, message)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error inserting post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/post.html"))
	tmpl.Execute(w, data)
}

func getCategoryIDByName(categoryName string) (int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var categoryID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = ?", categoryName).Scan(&categoryID)
	if err != nil {
		return 0, err
	}

	return categoryID, nil
}

func categories(w http.ResponseWriter, r *http.Request) {
	categories, err := getCategories()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/categories.html"))
	tmpl.Execute(w, categories)
}

func inside(w http.ResponseWriter, r *http.Request) {
	categories, err := getPost()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/inside.html"))
	tmpl.Execute(w, categories)
}

func create(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("TopicName")
		description := r.FormValue("Description")

		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Println(err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec("INSERT INTO categories (name, description) VALUES (?, ?)", name, description)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error inserting category", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

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

func getCategories() ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		categories = append(categories, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func getPost() ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		posts = append(posts, content)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// Function to handle the dynamic category pages
func categoryPageHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the category name from the URL
	categoryName := r.URL.Path[len("/categories/"):]

	// Create a template for the category page
	categoryTmpl := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>{{.}}</title>
		</head>
		<body>
			<h1>Welcome to the {{.}} category page!</h1>
		</body>
		</html>
	`
	// Parse and execute the template
	tmpl := template.Must(template.New("category").Parse(categoryTmpl))
	tmpl.Execute(w, categoryName)
}

func main() {
	SetupDatabase()
	SetupDatabase2()
	SetupDatabasePost()

	// Handle static files (CSS, JS, images, etc.)
	fs := http.FileServer(http.Dir("src"))
	http.Handle("/src/", http.StripPrefix("/src/", fs))

	http.HandleFunc("/categories/", categoryPageHandler)
	http.Handle("/home", http.HandlerFunc(landing))
	http.Handle("/auth", http.HandlerFunc(auth))
	http.Handle("/post", http.HandlerFunc(post))
	http.Handle("/categories", http.HandlerFunc(categories))
	http.Handle("/inside", http.HandlerFunc(inside))
	http.Handle("/about", http.HandlerFunc(about))
	http.Handle("/create", http.HandlerFunc(create))
	http.Handle("/user", http.HandlerFunc(user))
	http.Handle("/parameter", http.HandlerFunc(parameter))

	fmt.Println("Server is starting at http://localhost:8080")
	err := Open("http://localhost:8080/home")
	if err != nil {
		log.Printf("Failed to open URL: %v\n", err)
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
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

func SetupDatabase2() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL,
        description TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database setup completed.")
}

func SetupDatabasePost() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(category_id) REFERENCES categories(id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Post database setup completed.")
}
