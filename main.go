package main

import (
	"Forum/data"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	data.SetupDatabase()
	data.SetupDatabase2()
	data.SetupDatabasePost()
	data.SetupDatabaseCommentary()

	fs := http.FileServer(http.Dir("src"))
	http.Handle("/src/", http.StripPrefix("/src/", fs))

	http.Handle("/categories/", http.HandlerFunc(data.Categopost))
	http.Handle("/profile/", http.HandlerFunc(data.OtherProfile))
	http.Handle("/home", http.HandlerFunc(data.Landing))
	http.Handle("/auth", http.HandlerFunc(data.Auth))
	http.Handle("/post", http.HandlerFunc(data.Post))
	http.Handle("/categories", http.HandlerFunc(data.Categories))
	http.Handle("/inside", http.HandlerFunc(data.Inside))
	http.Handle("/about", http.HandlerFunc(data.About))
	http.Handle("/create", http.HandlerFunc(data.Create))
	http.Handle("/user", http.HandlerFunc(data.Users))
	http.Handle("/parameter", http.HandlerFunc(data.Parameter))

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
