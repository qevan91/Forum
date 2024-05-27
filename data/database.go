package data

import (
	"database/sql"
	"fmt"
	"log"
)

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

func SetupDatabaseCommentary() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS commentaries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		postID INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(postID) REFERENCES posts(id)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Commentary database setup completed.")
}
