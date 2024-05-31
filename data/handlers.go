package data

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

var errorAuth bool = false
var errorAbout bool = false
var errorCreate bool = false
var errorParameter bool = false

func Auth(w http.ResponseWriter, r *http.Request) {
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

func Landing(w http.ResponseWriter, r *http.Request) {
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

func Post(w http.ResponseWriter, r *http.Request) {
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

		var user_ID int
		err = db.QueryRow("SELECT id FROM users WHERE username = ?", session.Username).Scan(&user_ID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO posts (user_id, category_id, content) VALUES (?, ?, ?)", user_ID, category, message)
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
func Categories(w http.ResponseWriter, r *http.Request) {
	categories, err := getCategories()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/categories.html"))
	tmpl.Execute(w, categories)
}

func Inside(w http.ResponseWriter, r *http.Request) {
	categories, err := getPost()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/inside.html"))
	tmpl.Execute(w, categories)
}

func Create(w http.ResponseWriter, r *http.Request) {
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

func About(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/about.html"))
	tmpl.Execute(w, errorAbout)
}

func User(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Println(err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	session, err := validateSession(r)
	if err != nil {
		http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
		return
	}

	username := session.Username

	var email string
	err = db.QueryRow("SELECT email FROM users WHERE username = ?", username).Scan(&email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving user email", http.StatusInternalServerError)
		return
	}

	currentUsername := session.Username
	newEmail := r.FormValue("email")
	newUsername := r.FormValue("username")

	if newEmail != "" {
		_, err = db.Exec("UPDATE users SET email = ? WHERE username = ?", newEmail, currentUsername)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error updating email", http.StatusInternalServerError)
			return
		}
	}

	if newUsername != "" {
		var existingUser string
		err = db.QueryRow("SELECT username FROM users WHERE username = ?", newUsername).Scan(&existingUser)
		if err == nil {
			http.Error(w, "Le nom d'utilisateur est déjà pris", http.StatusConflict)
			return
		} else if err != sql.ErrNoRows {
			log.Println(err)
			http.Error(w, "Erreur de serveur", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, currentUsername)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error updating username", http.StatusInternalServerError)
			return
		}

		session.Username = newUsername
		deleteSession(w, r)
		_, err = createSession(w, newUsername)
		if err != nil {
			http.Error(w, "Impossible de créer la session", http.StatusInternalServerError)
			return
		}
	}

	tmpl := template.Must(template.ParseFiles("src/templates/user.html"))
	data := struct {
		Username string
		Email    string
	}{
		Username: username,
		Email:    email,
	}
	tmpl.Execute(w, data)
}

func Parameter(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/parameter.html"))
	tmpl.Execute(w, errorParameter)
}

func Categopost(w http.ResponseWriter, r *http.Request) {
	categoryName := strings.TrimPrefix(r.URL.Path, "/categories/")
	log.Println("Category:", categoryName)

	posts, userIDs, err := getPostsByCategory(categoryName)
	if err != nil {
		log.Println("Error retrieving posts:", err)
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}

	var usernames []string
	for _, userID := range userIDs {
		username, err := getUsernameByPostID(userID)
		name := strings.Join(username, " ")
		if err != nil {
			log.Println("Error retrieving username:", err)
			http.Error(w, "Error retrieving username", http.StatusInternalServerError)
			return
		}
		usernames = append(usernames, name)
	}

	tmplPath := filepath.Join("src/templates/inside.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Category  string
		Posts     []string
		Usernames []string
	}{
		Category:  categoryName,
		Posts:     posts,
		Usernames: usernames,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
