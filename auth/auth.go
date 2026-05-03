package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type AuthService struct {
	sessions map[string]*session
	mu       sync.RWMutex
}

type session struct {
	username  string
	createdAt time.Time
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
}

func NewAuthService() *AuthService {
	return &AuthService{
		sessions: make(map[string]*session),
	}
}

func (a *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Simple hardcoded authentication for demo purposes
	if req.Username == "admin" && req.Password == "admin123" {
		sessionID := generateSessionID()
		a.sessions[sessionID] = &session{
			username:  req.Username,
			createdAt: time.Now(),
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   3600,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{Message: "Login successful"})
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func (a *AuthService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		a.mu.Lock()
		delete(a.sessions, cookie.Value)
		a.mu.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

func (a *AuthService) Authenticate(r *http.Request) bool {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return false
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	_, exists := a.sessions[cookie.Value]
	return exists
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}