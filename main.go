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

func landing(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/landing.html"))
	tmpl.Execute(w, errorlanding)
}

func main() {
	http.Handle("/home", http.HandlerFunc(landing))
	http.Handle("/login", http.HandlerFunc(login))
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
