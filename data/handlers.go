package data

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var errorAbout bool = false
var errorCreate bool = false

type AuthData struct {
	Username string
	Error    string
}

func Auth(w http.ResponseWriter, r *http.Request) {
	data := AuthData{}

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
			var userCount int
			err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
			if err != nil {
				http.Error(w, "Erreur de serveur", http.StatusInternalServerError)
				return
			}

			role := "Users"
			if userCount == 0 {
				role = "Administrators"
			}

			hashedPassword, err := hashPassword(password)
			if err != nil {
				log.Fatal(err)
			}

			_, err = db.Exec("INSERT INTO users (username, email, password, Role) VALUES (?, ?, ?, ?)", username, email, hashedPassword, role)
			if err != nil {
				data.Error = "Erreur lors de l'inscription"
				renderTemplate(w, data)
				return
			}

			_, err = createSession(w, username)
			if err != nil {
				http.Error(w, "Impossible de créer la session", http.StatusInternalServerError)
				return
			}
		} else if username != "" && email == "" && password != "" {
			var storedHashedPassword string
			err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedHashedPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					data.Error = "Identifiant incorrect"
					renderTemplate(w, data)
					return
				}
				http.Error(w, "Erreur de serveur", http.StatusInternalServerError)
				return
			}

			if !checkPasswordHash(password, storedHashedPassword) {
				data.Error = "Mot de passe incorrect"
				renderTemplate(w, data)
				return
			}

			_, err = createSession(w, username)
			if err != nil {
				http.Error(w, "Impossible de créer la session", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	renderTemplate(w, data)
}

func renderTemplate(w http.ResponseWriter, data AuthData) {
	tmpl := template.Must(template.ParseFiles("src/templates/auth.html"))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

type LandingData struct {
	Username     string
	Error        string
	Usernames    []string
	Categories   []string
	Posts        []string
	PostID       []int
	UserID       []int
	ImagePaths   []string
	Dates        []string
	PostLikedIDs []int
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
		log.Println("Erreur lors de l'exécution du modèle:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

type FiltreData struct {
	Username           string
	Error              string
	Usernames          []string
	Categories         []string
	Posts              []string
	PostIDs            []int
	UserIDs            []int
	CurrentUserRole    string
	ImagePaths         []string
	Dates              []string
	PostLikedIDs       []int
	Commentaires       [][]string
	AuteurCommentaires [][]string
	LikeCounts         []int
	DislikeCounts      []int
}

func Filtre(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	session, err := validateSession(r)
	data := FiltreData{}

	if err != nil {
		data.Error = "Vous devez vous connecter"
	} else {
		data.Username = session.Username
		var currentUserRole string
		err = db.QueryRow("SELECT Role FROM users WHERE username = ?", session.Username).Scan(&currentUserRole)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error retrieving user role", http.StatusInternalServerError)
			return
		}
		data.CurrentUserRole = currentUserRole
	}

	categories, err := getCategories()
	if err != nil {
		log.Println("Erreur lors de la récupération des catégories:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	data.Categories = categories

	if r.Method == "POST" {
		categoryName := r.FormValue("categories")
		likedbut := r.FormValue("like")
		date := r.FormValue("date")

		if categoryName != "" {
			posts, postIDs, userIDs, imagePaths, dates, err := getPostsByCategory(categoryName)
			if err != nil {
				log.Println("Erreur lors de la récupération des articles par catégorie:", err)
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
				return
			}

			usernamepost, err := getUsernameByUserID(userIDs)
			if err != nil {
				log.Println("Erreur lors de la récupération des noms d'utilisateur:", err)
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
				return
			}

			likeCounts := make([]int, len(postIDs))
			dislikeCounts := make([]int, len(postIDs))
			for i, postID := range postIDs {
				likes, dislikes, err := getReactionsByPost(postID)
				if err != nil {
					log.Println("Error retrieving reactions for post ID:", postID, err)
					http.Error(w, "Error retrieving reactions", http.StatusInternalServerError)
					return
				}
				likeCounts[i] = likes
				dislikeCounts[i] = dislikes
			}

			localDates := make([]string, len(dates))
			for i, date := range dates {
				localTime, err := time.Parse(time.RFC3339, date)
				if err != nil {
					log.Println("Error parsing date:", err)
					localDates[i] = date
				} else {
					localDates[i] = localTime.Format("02-01-2006 15:04:05")
				}
			}

			data.Posts = posts
			data.PostIDs = postIDs
			data.UserIDs = userIDs
			data.Usernames = usernamepost
			data.ImagePaths = imagePaths
			data.Dates = localDates
			data.LikeCounts = likeCounts
			data.DislikeCounts = dislikeCounts
			var allCommentaires [][]string
			var allAuteurCommentaires [][]string

			for _, postID := range postIDs {
				comments, err := getComByPostID(postID)
				if err != nil {
					log.Println("Error retrieving comments for post ID:", postID, err)
					http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
					return
				}

				commentUserIDs, err := getUserByCom(postID)
				if err != nil {
					log.Println("Error retrieving authors for post ID:", postID, err)
					http.Error(w, "Error retrieving authors", http.StatusInternalServerError)
					return
				}

				var commentUsernames []string
				for _, userID := range commentUserIDs {
					username, err := getUsernameByPostID(userID)
					if err != nil {
						log.Println("Error retrieving username:", err)
						http.Error(w, "Error retrieving username", http.StatusInternalServerError)
						return
					}
					commentUsernames = append(commentUsernames, strings.Join(username, " "))
				}

				allCommentaires = append(allCommentaires, comments)
				allAuteurCommentaires = append(allAuteurCommentaires, commentUsernames)
			}
			data.Commentaires = allCommentaires
			data.AuteurCommentaires = allAuteurCommentaires
		} else if likedbut != "" || date == "" {

			posts, postIDs, userIDs, imagePaths, dates, err := getPost()
			if err != nil {
				log.Println("Erreur lors de la récupération des articles par catégorie:", err)
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
				return
			}

			localDates := make([]string, len(dates))
			for i, date := range dates {
				localTime, err := time.Parse(time.RFC3339, date)
				if err != nil {
					log.Println("Error parsing date:", err)
					localDates[i] = date
				} else {
					localDates[i] = localTime.Format("02-01-2006 15:04:05")
				}
			}

			usernamepost, err := getUsernameByUserID(userIDs)
			if err != nil {
				log.Println("Erreur lors de la récupération des noms d'utilisateur:", err)
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
				return
			}

			likeCounts := make([]int, len(postIDs))
			dislikeCounts := make([]int, len(postIDs))
			for i, postID := range postIDs {
				likes, dislikes, err := getReactionsByPost(postID)
				if err != nil {
					log.Println("Error retrieving reactions for post ID:", postID, err)
					http.Error(w, "Error retrieving reactions", http.StatusInternalServerError)
					return
				}
				likeCounts[i] = likes
				dislikeCounts[i] = dislikes
			}

			var allCommentaires [][]string
			var allAuteurCommentaires [][]string

			for _, postID := range postIDs {
				comments, err := getComByPostID(postID)
				if err != nil {
					log.Println("Error retrieving comments for post ID:", postID, err)
					http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
					return
				}

				commentUserIDs, err := getUserByCom(postID)
				if err != nil {
					log.Println("Error retrieving authors for post ID:", postID, err)
					http.Error(w, "Error retrieving authors", http.StatusInternalServerError)
					return
				}

				var commentUsernames []string
				for _, userID := range commentUserIDs {
					username, err := getUsernameByPostID(userID)
					if err != nil {
						log.Println("Error retrieving username:", err)
						http.Error(w, "Error retrieving username", http.StatusInternalServerError)
						return
					}
					commentUsernames = append(commentUsernames, strings.Join(username, " "))
				}

				allCommentaires = append(allCommentaires, comments)
				allAuteurCommentaires = append(allAuteurCommentaires, commentUsernames)
			}

			data.Posts = posts
			data.PostIDs = postIDs
			data.UserIDs = userIDs
			data.Usernames = usernamepost
			data.ImagePaths = imagePaths
			data.Dates = localDates
			data.LikeCounts = likeCounts
			data.DislikeCounts = dislikeCounts
			data.Commentaires = allCommentaires
			data.AuteurCommentaires = allAuteurCommentaires
		}
	}

	postLikedIDs, err := getLikedPost()
	if err != nil {
		log.Println("Erreur lors de la récupération des articles aimés:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	data.PostLikedIDs = postLikedIDs

	tmpl := template.Must(template.ParseFiles("src/templates/filtre.html"))
	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Erreur lors de l'exécution du modèle:", err)
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

	if r.Method == "POST" {
		category := r.FormValue("categories")
		message := r.FormValue("message")

		session, err := validateSession(r)
		if err != nil {
			http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Println(err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var userID int
		err = db.QueryRow("SELECT id FROM users WHERE username = ?", session.Username).Scan(&userID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
			return
		}

		err = os.MkdirAll("uploads", os.ModePerm)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil && err != http.ErrMissingFile {
			log.Println(err)
			http.Error(w, "Error retrieving file", http.StatusInternalServerError)
			return
		}

		var filepath string
		if err != http.ErrMissingFile {
			defer file.Close()

			if header.Size > 20*1024*1024 {
				http.Error(w, "File too large", http.StatusBadRequest)
				return
			}

			fileType := header.Header.Get("Content-Type")
			if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/gif" {
				http.Error(w, "Unsupported file type", http.StatusBadRequest)
				return
			}

			filename := fmt.Sprintf("%d-%s", userID, header.Filename)
			filepath = fmt.Sprintf("uploads/%s", filename)
			outFile, err := os.Create(filepath)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, file)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error writing file", http.StatusInternalServerError)
				return
			}
		}

		_, err = db.Exec("INSERT INTO posts (user_id, category_id, content, image_path) VALUES (?, ?, ?, ?)", userID, category, message, filepath)
		if err != nil {
			log.Println("Error inserting post:", err)
			http.Error(w, "Error inserting post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	data := struct {
		Categories []string
	}{
		Categories: categories,
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

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	session, err := validateSession(r)
	var currentUserRole string
	if err != nil {
		currentUserRole = "Guest"
	} else {
		err = db.QueryRow("SELECT Role FROM users WHERE username = ?", session.Username).Scan(&currentUserRole)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error retrieving user role", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "POST" && currentUserRole == "Administrators" {
		categoryID := r.FormValue("category-id")
		delete := r.FormValue("_method")

		if delete == "DELETE" {
			_, err = db.Exec("DELETE FROM categories WHERE name = ?", categoryID)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error deleting category", http.StatusInternalServerError)
				return
			}

			_, err = db.Exec("DELETE FROM posts WHERE category_id = ?", categoryID)
			if err != nil {
				http.Error(w, "Error deleting post", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/categories", http.StatusSeeOther)
		return
	}

	data := struct {
		Category        []string
		CurrentUserRole string
	}{
		Category:        categories,
		CurrentUserRole: currentUserRole,
	}

	tmpl := template.Must(template.ParseFiles("src/templates/categories.html"))
	tmpl.Execute(w, data)
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

type UserData struct {
	Username        string
	Email           string
	Posts           []string
	Commentaires    []string
	Date            []string
	Auth            string
	Liked           []string
	Disliked        []string
	IsAuthenticated bool
}

func Users(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Println(err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	session, err := validateSession(r)
	if err != nil {
		data := UserData{
			Auth:            "You have to be connected to see your profile",
			IsAuthenticated: false,
		}
		tmpl := template.Must(template.ParseFiles("src/templates/user.html"))
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
		}
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

	userID, dates, err := getUserByName(username)
	if err != nil {
		http.Error(w, "Error retrieving user data", http.StatusInternalServerError)
		return
	}

	coms, err := getComByUserID(userID)
	if err != nil {
		http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
		return
	}

	posts, err := getPostByUserID(userID)
	if err != nil {
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}

	localDates := make([]string, len(dates))
	for i, date := range dates {
		localTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			log.Println("Error parsing date:", err)
			localDates[i] = date
		} else {
			localDates[i] = localTime.Format("02-01-2006 15:04:05")
		}
	}

	postLikedIDs, err := getLikedPostByID(userID)
	if err != nil {
		log.Println("error accessing liked post", err)
		return
	}

	postDislikedIDs, err := getDislikedPost(userID)
	if err != nil {
		log.Println("Mes grosses couilles parties 1")
		return
	}

	var like []string
	var dislike []string
	for _, postid := range postLikedIDs {
		content, err := getPostByID(postid)
		if err != nil {
			log.Println("Mes grosses couilles")
			return
		}
		like = append(like, content...)
	}
	for _, postid := range postDislikedIDs {
		content, err := getPostByID(postid)
		if err != nil {
			log.Println("Mes grosses couilles")
			return
		}
		dislike = append(like, content...)
	}

	if r.Method == "POST" {
		newEmail := r.FormValue("email")
		newUsername := r.FormValue("username")
		newPassword := r.FormValue("password")

		if newEmail != "" {
			_, err = db.Exec("UPDATE users SET email = ? WHERE username = ?", newEmail, username)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error updating email", http.StatusInternalServerError)
				return
			}
		}

		if newPassword != "" {
			hashedPassword, err := hashPassword(newPassword)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				return
			}

			_, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", hashedPassword, username)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error updating password", http.StatusInternalServerError)
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

			_, err = db.Exec("UPDATE users SET username = ? WHERE username = ?", newUsername, username)
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

		if newUsername == "" && newEmail == "" && newPassword == "" {
			deleteSession(w, r)
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/user", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("src/templates/user.html"))
	data := UserData{
		Username:        username,
		Email:           email,
		Posts:           posts,
		Commentaires:    coms,
		Date:            localDates,
		Auth:            "",
		Liked:           like,
		Disliked:        dislike,
		IsAuthenticated: true,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
	}
}

func Parameter(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	session, err := validateSession(r)
	var currentUserRole string
	if err != nil {
		currentUserRole = "Guest"
	} else {
		err = db.QueryRow("SELECT Role FROM users WHERE username = ?", session.Username).Scan(&currentUserRole)
		if err != nil {
			http.Error(w, "Error retrieving user role", http.StatusInternalServerError)
			return
		}
	}

	var userID int
	err = db.QueryRow("SELECT ID FROM users WHERE username = ?", session.Username).Scan(&userID)
	if err != nil {
		http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
		return
	}

	userIDs, err := getUserIDFromRequest()
	if err != nil {
		http.Error(w, "Error retrieving user IDs", http.StatusInternalServerError)
		return
	}

	names, err := getUsernameByUserID(userIDs)
	if err != nil {
		http.Error(w, "Error retrieving usernames", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if currentUserRole == "Users" {
			var exists int
			err = db.QueryRow("SELECT COUNT(*) FROM request WHERE user_id = ?", userID).Scan(&exists)
			if err != nil {
				http.Error(w, "Error checking request", http.StatusInternalServerError)
				return
			}

			if exists > 0 {
				http.Error(w, "You have already applied to become a moderator", http.StatusForbidden)
			} else {
				_, err = db.Exec("INSERT INTO request (user_id) VALUES (?)", userID)
				if err != nil {
					http.Error(w, "Error inserting request", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}
		} else if currentUserRole == "Administrators" {
			requestedUserID := r.FormValue("userID")
			delete := r.FormValue("Delete")

			if delete != "" {
				_, err = db.Exec("DELETE FROM request WHERE user_id = ?", requestedUserID)
				if err != nil {
					http.Error(w, "Error deleting request", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/parameter", http.StatusSeeOther)
				return
			}
		}
	}

	data := struct {
		CurrentUserRole string
		Names           []string
		UserIDs         []int
	}{
		CurrentUserRole: currentUserRole,
		Names:           names,
		UserIDs:         userIDs,
	}

	tmpl := template.Must(template.ParseFiles("src/templates/parameter.html"))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func Categopost(w http.ResponseWriter, r *http.Request) {
	categoryName := strings.TrimPrefix(r.URL.Path, "/categories/")
	posts, postIDs, userIDs, imagePaths, dates, err := getPostsByCategory(categoryName)
	if err != nil {
		log.Println("Error retrieving posts:", err)
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}

	session, err := validateSession(r)
	var currentUserRole string
	if err != nil {
		currentUserRole = "Guest"
	} else {
		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Println("Database connection error:", err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		err = db.QueryRow("SELECT Role FROM users WHERE username = ?", session.Username).Scan(&currentUserRole)
		if err != nil {
			log.Println("Error retrieving user role:", err)
			http.Error(w, "Error retrieving user role", http.StatusInternalServerError)
			return
		}
	}

	localDates := make([]string, len(dates))
	for i, date := range dates {
		localTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			log.Println("Error parsing date:", err)
			localDates[i] = date
		} else {
			localDates[i] = localTime.Format("02-01-2006 15:04:05")
		}
	}

	var usernames []string
	for _, userID := range userIDs {
		username, err := getUsernameByPostID(userID)
		if err != nil {
			log.Println("Error retrieving username:", err)
			http.Error(w, "Error retrieving username", http.StatusInternalServerError)
			return
		}
		usernames = append(usernames, strings.Join(username, " "))
	}

	var allCommentaires [][]string
	var allAuteurCommentaires [][]string

	for _, postID := range postIDs {
		comments, err := getComByPostID(postID)
		if err != nil {
			log.Println("Error retrieving comments for post ID:", postID, err)
			http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
			return
		}

		commentUserIDs, err := getUserByCom(postID)
		if err != nil {
			log.Println("Error retrieving authors for post ID:", postID, err)
			http.Error(w, "Error retrieving authors", http.StatusInternalServerError)
			return
		}

		var commentUsernames []string
		for _, userID := range commentUserIDs {
			username, err := getUsernameByPostID(userID)
			if err != nil {
				log.Println("Error retrieving username:", err)
				http.Error(w, "Error retrieving username", http.StatusInternalServerError)
				return
			}
			commentUsernames = append(commentUsernames, strings.Join(username, " "))
		}

		allCommentaires = append(allCommentaires, comments)
		allAuteurCommentaires = append(allAuteurCommentaires, commentUsernames)
	}

	likeCounts := make([]int, len(postIDs))
	dislikeCounts := make([]int, len(postIDs))
	for i, postID := range postIDs {
		likes, dislikes, err := getReactionsByPost(postID)
		if err != nil {
			log.Println("Error retrieving reactions for post ID:", postID, err)
			http.Error(w, "Error retrieving reactions", http.StatusInternalServerError)
			return
		}
		likeCounts[i] = likes
		dislikeCounts[i] = dislikes
	}

	if r.Method == "POST" {
		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Println("Database connection error:", err)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		content := r.FormValue("reply-message")
		postID := r.FormValue("post-id")

		if content != "" {
			var userID int
			err = db.QueryRow("SELECT id FROM users WHERE username = ?", session.Username).Scan(&userID)
			if err != nil {
				log.Println("Error retrieving user ID:", err)
				http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
				return
			}
			_, err = db.Exec("INSERT INTO commentaries (postID, user_ID, content) VALUES (?, ?, ?)", postID, userID, content)
			if err != nil {
				http.Error(w, "Error inserting commentary", http.StatusInternalServerError)
				return
			}
		}

		if currentUserRole != "Guest" {

			reactionStr := r.FormValue("reaction")

			if reactionStr != "" {
				reaction, err := strconv.Atoi(reactionStr)
				if err != nil {
					http.Error(w, "Invalid reaction value", http.StatusBadRequest)
					return
				}

				var userID int
				err = db.QueryRow("SELECT id FROM users WHERE username = ?", session.Username).Scan(&userID)
				if err != nil {
					log.Println("Error retrieving user ID:", err)
					http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
					return
				}

				var existingReaction int
				err = db.QueryRow("SELECT reaction FROM reactions WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&existingReaction)

				if err == sql.ErrNoRows {
					_, err = db.Exec("INSERT INTO reactions (post_id, user_id, reaction) VALUES (?, ?, ?)", postID, userID, reaction)
				} else if err == nil {
					_, err = db.Exec("UPDATE reactions SET reaction = ? WHERE post_id = ? AND user_id = ?", reaction, postID, userID)
				}

				if err != nil {
					log.Println("Error handling reaction:", err)
					http.Error(w, "Error handling reaction", http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			}

			action := r.FormValue("_method")

			if action == "DELETE" && (currentUserRole == "Administrators" || currentUserRole == "Moderators") {
				_, err = db.Exec("DELETE FROM posts WHERE ID = ?", postID)
				if err != nil {
					log.Println("Error deleting post:", err)
					http.Error(w, "Error deleting post", http.StatusInternalServerError)
					return
				}
			} else if action == "REPORT" {
				_, err = db.Exec("UPDATE posts SET signaler = TRUE WHERE ID = ?", postID)
				if err != nil {
					log.Println("Error reporting post:", err)
					http.Error(w, "Error reporting post", http.StatusInternalServerError)
					return
				}
			}

			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		} else {
			http.Error(w, "Unauthorized action", http.StatusUnauthorized)
			return
		}
	}

	tmplPath := filepath.Join("src/templates/categopost.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Category           string
		Posts              []string
		PostIDs            []int
		Usernames          []string
		Commentaires       [][]string
		AuteurCommentaires [][]string
		ImagePaths         []string
		LikeCounts         []int
		DislikeCounts      []int
		CurrentUserRole    string
		LocalDates         []string
	}{
		Category:           categoryName,
		Posts:              posts,
		PostIDs:            postIDs,
		Usernames:          usernames,
		Commentaires:       allCommentaires,
		AuteurCommentaires: allAuteurCommentaires,
		ImagePaths:         imagePaths,
		LikeCounts:         likeCounts,
		DislikeCounts:      dislikeCounts,
		CurrentUserRole:    currentUserRole,
		LocalDates:         localDates,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func OtherProfile(w http.ResponseWriter, r *http.Request) {
	OtUsername := strings.TrimPrefix(r.URL.Path, "/profile/")

	id, dates, err := getUserByName(OtUsername)
	if err != nil {
		log.Println("Error retrieving user:", err)
		http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		return
	}

	content, err := getPostByID(id)
	if err != nil {
		log.Println("Error retrieving posts:", err)
		http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
		return
	}

	localDates := make([]string, len(dates))
	for i, date := range dates {
		localTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			log.Println("Error parsing date:", err)
			localDates[i] = date
		} else {
			localDates[i] = localTime.Format("02-01-2006 15:04:05")
		}
	}

	session, err := validateSession(r)
	if err != nil {
		http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
		return
	}

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var currentUserRole string
	err = db.QueryRow("SELECT Role FROM users WHERE username = ?", session.Username).Scan(&currentUserRole)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving user role", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		role := r.FormValue("role")

		validRoles := GetRoles()
		isValidRole := false
		for _, r := range validRoles {
			if role == r {
				isValidRole = true
				break
			}
		}

		if !isValidRole {
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}

		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec("UPDATE users SET Role = ? WHERE username = ?", role, OtUsername)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error updating role", http.StatusInternalServerError)
			return
		}
	}

	roles := GetRoles()

	tmpl := template.Must(template.ParseFiles("src/templates/otherprofile.html"))
	data := struct {
		Dates           []string
		Post            []string
		Username        string
		Status          []string
		CurrentUserRole string
	}{
		Dates:           localDates,
		Post:            content,
		Username:        OtUsername,
		Status:          roles,
		CurrentUserRole: currentUserRole,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors du rendu de la page", http.StatusInternalServerError)
	}
}

func Panel(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	session, err := validateSession(r)
	var currentUserRole string
	var userID int

	if err != nil {
		currentUserRole = "Guest"
	} else {
		err = db.QueryRow("SELECT Role, ID FROM users WHERE username = ?", session.Username).Scan(&currentUserRole, &userID)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération du rôle ou de l'ID de l'utilisateur", http.StatusInternalServerError)
			return
		}
	}

	post, postid, PostuserId, image, date, err := getReportedPost()
	if err != nil {
		log.Println("Error accessing reported posts")
		log.Println(date)
		log.Println(postid)
		return
	}

	var usernames []string
	for _, userID := range PostuserId {
		username, err := getUsernameByPostID(userID)
		if err != nil {
			log.Println("Error retrieving username:", err)
			http.Error(w, "Error retrieving username", http.StatusInternalServerError)
			return
		}
		usernames = append(usernames, strings.Join(username, " "))
	}

	if r.Method == "POST" {
		if currentUserRole == "Users" {
			var exists int
			err := db.QueryRow("SELECT COUNT(*) FROM request WHERE user_id = ?", userID).Scan(&exists)
			if err != nil {
				http.Error(w, "Erreur lors de la vérification de la demande", http.StatusInternalServerError)
				return
			}

			if exists > 0 {
				http.Error(w, "Vous avez déjà postulé pour devenir modérateur", http.StatusForbidden)
			} else {
				_, err = db.Exec("INSERT INTO request (user_id) VALUES (?)", userID)
				if err != nil {
					http.Error(w, "Erreur lors de l'insertion de la demande", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}
		} else if currentUserRole == "Administrators" || currentUserRole == "Moderators" {
			requestedUserID := r.FormValue("userID")
			delete := r.FormValue("Delete")
			action := r.FormValue("_method")
			postID := r.FormValue("post-id")

			if delete != "" {
				_, err := db.Exec("DELETE FROM request WHERE user_id = ?", requestedUserID)
				if err != nil {
					http.Error(w, "Erreur lors de la suppression de la demande", http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/panel", http.StatusSeeOther)
				return
			}

			if action == "DELETE" && postID != "" {
				_, err := db.Exec("DELETE FROM posts WHERE id = ?", postID)
				if err != nil {
					http.Error(w, "Erreur lors de la suppression du post", http.StatusInternalServerError)
					return
				}
				// Redirection après suppression
				http.Redirect(w, r, "/panel", http.StatusSeeOther)
				return
			}
		}
	}

	userIDs, err := getUserIDFromRequest()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des IDs utilisateurs", http.StatusInternalServerError)
		return
	}

	names, err := getUsernameByUserID(userIDs)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des noms d'utilisateurs", http.StatusInternalServerError)
		return
	}

	data := struct {
		CurrentUserRole string
		Names           []string
		UserIDs         []int
		PostIDs         []int
		Usernames       []string
		Posts           []string
		ImagePaths      []string
	}{
		CurrentUserRole: currentUserRole,
		Names:           names,
		UserIDs:         userIDs,
		PostIDs:         postid,
		Usernames:       usernames,
		Posts:           post,
		ImagePaths:      image,
	}

	tmpl, err := template.ParseFiles("src/templates/panel.html")
	if err != nil {
		http.Error(w, "Erreur lors du rendu du template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors de l'exécution du template", http.StatusInternalServerError)
	}
}
