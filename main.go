package main

import (
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

var (
	errorLogin   bool = false
	errorLanding bool = false
)

// Gestionnaire pour la page de login
func loginHandler(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("src/templates/landing_n.html")
	if err != nil {
		log.Println("Erreur lors du chargement du template de login:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, errorLogin); err != nil {
		log.Println("Erreur lors de l'exécution du template de login:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// Gestionnaire pour la page d'accueil
func landingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")

	tmpl, err := template.ParseFiles("src/templates/landing.html")
	if err != nil {
		log.Println("Erreur lors du chargement du template de landing:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, errorLanding); err != nil {
		log.Println("Erreur lors de l'exécution du template de landing:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// Gestionnaire pour les fichiers statiques (comme les fichiers CSS)
func staticHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header to 'text/css' for CSS files
	w.Header().Set("Content-Type", "text/css")

	// Serve the static file
	http.ServeFile(w, r, r.URL.Path[1:])
}

// Fonction pour ouvrir une URL dans le navigateur
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
	// Gestion des routes
	http.HandleFunc("/src/css/", staticHandler)
	http.HandleFunc("/home", loginHandler)

	// Serveur HTTP
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Erreur lors du démarrage du serveur HTTP:", err)
		}
	}()

	// Ouvrir le navigateur avec l'URL d'accueil
	if err := Open("http://localhost:8080/home"); err != nil {
		log.Println("Erreur lors de l'ouverture du navigateur:", err)
	}

	// Maintenir le programme actif
	select {}
}
