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

func getUserByName(Username string) (int, []string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return 0, nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, created_at FROM users WHERE username = ?", Username)
	var id int
	var date string
	if err := row.Scan(&id, &date); err != nil {
		return 0, nil, err
	}
	dates := []string{date}
	return id, dates, nil
}

func getPostByID(postID int) ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM posts WHERE user_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []string
	for rows.Next() {
		var post string
		if err := rows.Scan(&post); err != nil {
			return nil, err
		}
		content = append(content, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return content, nil
}
