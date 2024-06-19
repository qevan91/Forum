package data

import (
	"database/sql"
	"fmt"
	"strings"
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

func getPostByUserID(userID int) ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM posts WHERE user_id = ?", userID)
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

func getPostsByCategory(categoryName string) ([]string, []int, []int, []string, []string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, content, user_id, image_path, created_at FROM posts WHERE category_id = ?", categoryName)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	var posts []string
	var postIDs []int
	var userIDs []int
	var imagePaths []string
	var dates []string
	for rows.Next() {
		var id int
		var content string
		var userID int
		var imagePath string
		var date string
		if err := rows.Scan(&id, &content, &userID, &imagePath, &date); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		postIDs = append(postIDs, id)
		posts = append(posts, content)
		userIDs = append(userIDs, userID)
		imagePaths = append(imagePaths, imagePath)
		dates = append(dates, date)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return posts, postIDs, userIDs, imagePaths, dates, nil
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

func getComByUserID(userID int) ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM commentaries WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []string
	for rows.Next() {
		var com string
		if err := rows.Scan(&com); err != nil {
			return nil, err
		}
		content = append(content, com)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return content, nil
}

func getComByPostID(postID int) ([]string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM commentaries WHERE postID = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content []string
	for rows.Next() {
		var com string
		if err := rows.Scan(&com); err != nil {
			return nil, err
		}
		content = append(content, com)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return content, nil
}

func getUserByCom(postID int) ([]int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT user_id FROM commentaries WHERE postID = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

func getReactionsByPost(postID int) (int, int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return 0, 0, err
	}
	defer db.Close()

	var likes, dislikes int
	query := `
		SELECT
			COUNT(CASE WHEN reaction = 1 THEN 1 END) as likes,
			COUNT(CASE WHEN reaction = -1 THEN 1 END) as dislikes
		FROM reactions
		WHERE post_id = ?`
	err = db.QueryRow(query, postID).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}

func getUserIDFromRequest() ([]int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT user_id FROM request")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func getUsernameByUserID(userIDs []int) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := "SELECT Username FROM users WHERE ID IN (" + strings.Repeat("?,", len(userIDs)-1) + "?)"

	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	rows, err := db.Query(query, args...)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usernames, nil
}

func getLikedPost(userID int) ([]int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT post_id FROM reactions WHERE user_id = ? AND reaction = 1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			return nil, err
		}
		posts = append(posts, postID)
	}

	return posts, nil
}

func getDislikedPost(userID int) ([]int, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT post_id FROM reactions WHERE user_id = ? AND reaction = -1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			return nil, err
		}
		posts = append(posts, postID)
	}

	return posts, nil
}

func getReportedPost() ([]string, []int, []int, []string, []string, error) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, content, user_id, image_path, created_at FROM posts WHERE signaler = ?", 1)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	var posts []string
	var postIDs []int
	var userIDs []int
	var imagePaths []string
	var dates []string

	for rows.Next() {
		var id int
		var content string
		var userID int
		var imagePath string
		var date string

		if err := rows.Scan(&id, &content, &userID, &imagePath, &date); err != nil {
			return nil, nil, nil, nil, nil, err
		}

		postIDs = append(postIDs, id)
		posts = append(posts, content)
		userIDs = append(userIDs, userID)
		imagePaths = append(imagePaths, imagePath)
		dates = append(dates, date)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return posts, postIDs, userIDs, imagePaths, dates, nil
}
