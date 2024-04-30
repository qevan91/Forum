package main

import (
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
)

var errorLogin bool = false
var errorlanding bool = false

func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/login.html"))
	tmpl.Execute(w, errorLogin)
}

/*
	func loginHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed) // Code 405
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")


		isValid, err := fonction.ValidateLogin(username, password)
		if err != nil {
			http.Error(w, "Erreur lors de la validation du login", http.StatusInternalServerError)
			return
		}

		if isValid {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		} else {
			tmpl := template.Must(template.ParseFiles("src/templates/login.html"))
			tmpl.Execute(w, map[string]bool{"Error": true})
		}
	}
*/
func landing(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/landing.html"))
	tmpl.Execute(w, errorlanding)
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
	http.Handle("/home", http.HandlerFunc(landing))
	//http.HandleFunc("/login", loginHandler)
	http.Handle("/login", http.HandlerFunc(login))
	Open("http://localhost/home")
	http.ListenAndServe("", nil)
}
