package data

import "database/sql"

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

func getPostsByCategory(categoryName string) ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM posts WHERE category_id = ?", categoryName)
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
