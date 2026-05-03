package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Prem5123/autodev-target/auth"
	"github.com/Prem5123/autodev-target/notes"
)

type Handler struct {
	noteService *notes.Service
	authService *auth.AuthService
}

type CreateNoteRequest struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Template string `json:"template"`
}

type UpdateNoteRequest struct {
	ID       string `json:"id,omitempty"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Template string `json:"template"`
}

func NewHandler(noteService *notes.Service, authService *auth.AuthService) *Handler {
	return &Handler{
		noteService: noteService,
		authService: authService,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Apply authentication middleware to all /api/notes routes
	notesHandler := h.authMiddleware(h.handleNotes)

	mux.HandleFunc("GET /api/notes", notesHandler)
	mux.HandleFunc("POST /api/notes", notesHandler)
	mux.HandleFunc("GET /api/notes/", notesHandler)
	mux.HandleFunc("PUT /api/notes/", notesHandler)
	mux.HandleFunc("DELETE /api/notes/", notesHandler)
}

func (h *Handler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.authService.Authenticate(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (h *Handler) handleNotes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// GET /api/notes - List all notes
	if r.Method == http.MethodGet && path == "/api/notes" {
		h.listNotes(w, r)
		return
	}

	// POST /api/notes - Create a note
	if r.Method == http.MethodPost && path == "/api/notes" {
		h.createNote(w, r)
		return
	}

	// GET /api/notes/:id - Get a specific note
	if r.Method == http.MethodGet && strings.HasPrefix(path, "/api/notes/") {
		id := strings.TrimPrefix(path, "/api/notes/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		h.getNote(w, r, id)
		return
	}

	// PUT /api/notes/:id - Update a note
	if r.Method == http.MethodPut && strings.HasPrefix(path, "/api/notes/") {
		id := strings.TrimPrefix(path, "/api/notes/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		h.updateNote(w, r, id)
		return
	}

	// DELETE /api/notes/:id - Delete a note
	if r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/notes/") {
		id := strings.TrimPrefix(path, "/api/notes/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		h.deleteNote(w, r, id)
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	allNotes, err := h.noteService.List()
	if err != nil {
		http.Error(w, "Failed to list notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allNotes)
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check for empty title and content
	if strings.TrimSpace(req.Title) == "" && strings.TrimSpace(req.Content) == "" {
		http.Error(w, "Title and content cannot both be empty", http.StatusBadRequest)
		return
	}

	note, err := h.noteService.Create(req.Title, req.Content, req.Template)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request, id string) {
	// Check for non-numeric id (simple validation)
	if _, err := strconv.Atoi(id); err != nil && !isValidUUID(id) {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	note, err := h.noteService.Get(id)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}
	if note == nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request, id string) {
	// Check for non-numeric id
	if _, err := strconv.Atoi(id); err != nil && !isValidUUID(id) {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check for mismatched URL id and body id
	if req.ID != "" && req.ID != id {
		http.Error(w, "URL id and body id mismatch", http.StatusBadRequest)
		return
	}

	// Check for empty title and content
	if strings.TrimSpace(req.Title) == "" && strings.TrimSpace(req.Content) == "" {
		http.Error(w, "Title and content cannot both be empty", http.StatusBadRequest)
		return
	}

	note, err := h.noteService.Update(id, req.Title, req.Content, req.Template)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request, id string) {
	// Check for non-numeric id
	if _, err := strconv.Atoi(id); err != nil && !isValidUUID(id) {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	err := h.noteService.Delete(id)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for _, c := range s {
		if c == '-' {
			continue
		}
		if c >= '0' && c <= '9' {
			continue
		}
		if c >= 'a' && c <= 'f' {
			continue
		}
		if c >= 'A' && c <= 'F' {
			continue
		}
		return false
	}
	return true
}