package data

import (
	"database/sql"
	"fmt"
)

func getCategories() ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		categories = append(categories, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func getPost() ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		posts = append(posts, content)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func getPostsByCategory(categoryName string) ([]string, []int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content, user_id FROM posts WHERE category_id = ?", categoryName)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var posts []string
	var userIDs []int
	for rows.Next() {
		var content string
		var userID int
		if err := rows.Scan(&content, &userID); err != nil {
			return nil, nil, err
		}
		posts = append(posts, content)
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return posts, userIDs, nil
}
func getUserIDByPost(post string) (int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT user_id FROM posts").Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("aucun utilisateur trouvé pour le post ID")
		}
		return 0, err
	}

	return userID, nil
}

func getUsernameByPostID(ID int) ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT username FROM users WHERE id = ?", ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(usernames) == 0 {
		return nil, fmt.Errorf("aucun utilisateur trouvé pour le post ID %d", ID)
	}

	return usernames, nil
}
