package main

import (
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
)

var errorAbout bool = false

func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/login.html"))
	tmpl.Execute(w, errorAbout)
}

func main() {
	//http.HandleFunc("/", about)
	http.Handle("/", http.HandlerFunc(login))
	Open("http://localhost/")
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
