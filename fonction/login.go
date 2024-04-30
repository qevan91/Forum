package fonction

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Fonction pour lire les utilisateurs à partir d'un fichier texte
func ReadUsersFromFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	users := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			continue
		}
		username := parts[0]
		password := parts[1]
		users[username] = password
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Fonction pour valider le login en comparant les mots de passe
func ValidateLogin(username, password string) (bool, error) {
	users, err := ReadUsersFromFile("users.txt")
	if err != nil {
		return false, err
	}

	storedPassword, exists := users[username]
	if !exists {
		fmt.Println("Nom d'utilisateur non trouvé:", username)
		return false, nil
	}

	isValid := storedPassword == password
	if isValid {
		fmt.Println("Connexion réussie pour:", username)
	} else {
		fmt.Println("Mot de passe incorrect pour:", username)
	}
	return isValid, nil
}
