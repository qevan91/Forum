package main

import (
	"Forum/data"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Database initialization
	data.SetupDatabase()
	data.SetupDatabase2()
	data.SetupDatabasePost()
	data.SetupDatabaseReactions()
	data.SetupDatabaseCommentary()
	data.SetupDatabaseRequestRole()

	fs := http.FileServer(http.Dir("src"))
	http.Handle("/src/", http.StripPrefix("/src/", fs))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	http.Handle("/categories/", http.HandlerFunc(data.Categopost))
	http.Handle("/profile/", http.HandlerFunc(data.OtherProfile))
	http.Handle("/home", http.HandlerFunc(data.Landing))
	http.Handle("/auth", http.HandlerFunc(data.Auth))
	http.Handle("/post", http.HandlerFunc(data.Post))
	http.Handle("/categories", http.HandlerFunc(data.Categories))
	http.Handle("/about", http.HandlerFunc(data.About))
	http.Handle("/create", http.HandlerFunc(data.Create))
	http.Handle("/user", http.HandlerFunc(data.Users))
	http.Handle("/parameter", http.HandlerFunc(data.Parameter))
	http.Handle("/panel", http.HandlerFunc(data.Panel))

	err := Open("http://localhost:8080/home")
	if err != nil {
		log.Printf("Failed to open URL: %v\n", err)
	}

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}

// Opens the default browser
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
