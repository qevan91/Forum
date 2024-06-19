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

// sessionStore is a thread-safe in-memory store for session data.
var sessionStore = struct {
	sync.RWMutex
	sessions map[string]Session
}{
	sessions: make(map[string]Session),
}

// generateSessionID generates a new secure random session ID.
func generateSessionID() (string, error) {
	b := make([]byte, 32)  // Create a byte slice with length 32
	_, err := rand.Read(b) // Fill the byte slice with random bytes
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// createSession creates a new session for the given username, stores it, and sets a cookie on the response.
func createSession(w http.ResponseWriter, username string) (string, error) {
	// Generate a new session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(24 * time.Hour)

	// Store the session in the session store
	sessionStore.Lock()
	sessionStore.sessions[sessionID] = Session{
		Username: username,
		Expiry:   expiry,
	}
	sessionStore.Unlock()

	// Set the session cookie in the HTTP response
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionID,
		Expires: expiry,
	})

	return sessionID, nil
}

// validateSession checks if the session from the request is valid and not expired.
func validateSession(r *http.Request) (*Session, error) {
	// Retrieve the session cookie from the request
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	// Read session data from the session store
	sessionStore.RLock()
	session, exists := sessionStore.sessions[cookie.Value]
	sessionStore.RUnlock()

	// Check if the session exists and has not expired
	if !exists || session.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("session invalid or expired")
	}

	return &session, nil
}

// deleteSession deletes the session associated with the request and removes the session cookie.
func deleteSession(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session cookie from the request
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return
	}

	// Remove the session from the session store
	sessionStore.Lock()
	delete(sessionStore.sessions, cookie.Value)
	sessionStore.Unlock()

	// Remove the session cookie by setting its MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})
}
