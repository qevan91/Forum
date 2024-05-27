package data

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	Username string
	Expiry   time.Time
}

var sessionStore = struct {
	sync.RWMutex
	sessions map[string]Session
}{
	sessions: make(map[string]Session),
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func createSession(w http.ResponseWriter, username string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(24 * time.Hour)

	sessionStore.Lock()
	sessionStore.sessions[sessionID] = Session{
		Username: username,
		Expiry:   expiry,
	}
	sessionStore.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionID,
		Expires: expiry,
	})

	return sessionID, nil
}

func validateSession(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	sessionStore.RLock()
	session, exists := sessionStore.sessions[cookie.Value]
	sessionStore.RUnlock()

	if !exists || session.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("session invalid or expired")
	}

	return &session, nil
}

func deleteSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return
	}

	sessionStore.Lock()
	delete(sessionStore.sessions, cookie.Value)
	sessionStore.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})
}
