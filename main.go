package main

import (
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
)

var errorLogin bool = false
var errorLanding bool = false
var errorPost bool = false
var errorCategories bool = false
var errorInside bool = false
var errorAbout bool = false
var errorCreate bool = false
var errorUser bool = false
var errorParameter bool = false

func login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/login.html"))
	tmpl.Execute(w, errorLogin)
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

/*func parameter(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/parameter.html"))
	tmpl.Execute(w, errorParameter)
}*/

func main() {
	http.Handle("/home", http.HandlerFunc(landing))
	http.Handle("/login", http.HandlerFunc(login))
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
