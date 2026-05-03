package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Prem5123/autodev-target/auth"
	"github.com/Prem5123/autodev-target/notes"
)

func setupTestEnv(t *testing.T) (*Handler, *notes.FileStorage, *http.ServeMux, *auth.AuthService, func()) {
	tmpFile, err := os.CreateTemp("", "notes_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// Initialize with empty JSON object since the storage doesn't handle empty files
	if err := os.WriteFile(tmpFile.Name(), []byte("{}"), 0644); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to initialize temp file: %v", err)
	}

	storage, err := notes.NewFileStorage(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create storage: %v", err)
	}

	noteService := notes.NewService(storage)
	authService := auth.NewAuthService()
	handler := NewHandler(noteService, authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", authService.LoginHandler)
	mux.HandleFunc("/api/auth/logout", authService.LogoutHandler)
	handler.RegisterRoutes(mux)

	cleanup := func() {
		storage.Close()
		os.Remove(tmpFile.Name())
	}

	return handler, storage, mux, authService, cleanup
}

func login(t *testing.T, authService *auth.AuthService) *http.Cookie {
	body := bytes.NewBufferString(`{"username":"admin","password":"admin123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	authService.LoginHandler(w, req)

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session_id" {
			return c
		}
	}
	return nil
}

func TestCreateAndListNote(t *testing.T) {
	_, _, mux, authService, cleanup := setupTestEnv(t)
	defer cleanup()

	sessionCookie := login(t, authService)

	t.Run("POST /api/notes creates a note", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title":"Test Note","content":"Test Content","template":"soap"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var note notes.Note
		if err := json.NewDecoder(w.Body).Decode(&note); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if note.ID == "" {
			t.Error("Expected non-empty note ID")
		}
		if note.Title != "Test Note" {
			t.Errorf("Expected title 'Test Note', got '%s'", note.Title)
		}
		if note.Content != "Test Content" {
			t.Errorf("Expected content 'Test Content', got '%s'", note.Content)
		}
		if note.Template != "soap" {
			t.Errorf("Expected template 'soap', got '%s'", note.Template)
		}
		if note.CreatedAt == "" {
			t.Error("Expected non-empty created_at")
		}
	})

	t.Run("GET /api/notes lists created note", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notes", nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var notes []notes.Note
		if err := json.NewDecoder(w.Body).Decode(&notes); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(notes) != 1 {
			t.Errorf("Expected 1 note, got %d", len(notes))
		}
		if notes[0].Title != "Test Note" {
			t.Errorf("Expected title 'Test Note', got '%s'", notes[0].Title)
		}
	})
}

func TestGetNoteByID(t *testing.T) {
	_, _, mux, authService, cleanup := setupTestEnv(t)
	defer cleanup()

	sessionCookie := login(t, authService)

	body := bytes.NewBufferString(`{"title":"Test Note","content":"Test Content","template":"soap"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var createdNote notes.Note
	json.NewDecoder(w.Body).Decode(&createdNote)

	t.Run("GET /api/notes/:id returns the note", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notes/"+createdNote.ID, nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var note notes.Note
		if err := json.NewDecoder(w.Body).Decode(&note); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if note.ID != createdNote.ID {
			t.Errorf("Expected ID '%s', got '%s'", createdNote.ID, note.ID)
		}
	})

	t.Run("GET /api/notes/:id with non-existent id returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notes/non-existent-id", nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("GET /api/notes/:id with non-numeric id returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notes/abc", nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestUpdateNote(t *testing.T) {
	_, _, mux, authService, cleanup := setupTestEnv(t)
	defer cleanup()

	sessionCookie := login(t, authService)

	body := bytes.NewBufferString(`{"title":"Original Title","content":"Original Content","template":"soap"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var createdNote notes.Note
	json.NewDecoder(w.Body).Decode(&createdNote)

	t.Run("PUT /api/notes/:id updates the note", func(t *testing.T) {
		updateBody := bytes.NewBufferString(`{"title":"Updated Title","content":"Updated Content","template":"soap"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/notes/"+createdNote.ID, updateBody)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var updatedNote notes.Note
		if err := json.NewDecoder(w.Body).Decode(&updatedNote); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if updatedNote.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", updatedNote.Title)
		}
		if updatedNote.Content != "Updated Content" {
			t.Errorf("Expected content 'Updated Content', got '%s'", updatedNote.Content)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/notes/"+createdNote.ID, nil)
		req.AddCookie(sessionCookie)

		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var persistedNote notes.Note
		json.NewDecoder(w.Body).Decode(&persistedNote)

		if persistedNote.Title != "Updated Title" {
			t.Errorf("Expected persisted title 'Updated Title', got '%s'", persistedNote.Title)
		}
	})

	t.Run("PUT /api/notes/:id with mismatched URL and body id returns 400", func(t *testing.T) {
		updateBody := bytes.NewBufferString(`{"id":"different-id","title":"Updated Title","content":"Updated Content","template":"soap"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/notes/"+createdNote.ID, updateBody)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestDeleteNote(t *testing.T) {
	_, _, mux, authService, cleanup := setupTestEnv(t)
	defer cleanup()

	sessionCookie := login(t, authService)

	body := bytes.NewBufferString(`{"title":"Test Note","content":"Test Content","template":"soap"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(sessionCookie)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var createdNote notes.Note
	json.NewDecoder(w.Body).Decode(&createdNote)

	t.Run("DELETE /api/notes/:id removes the note and returns 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/notes/"+createdNote.ID, nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/notes/"+createdNote.ID, nil)
		req.AddCookie(sessionCookie)

		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d after deletion, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("DELETE /api/notes/:id for non-existent id returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/notes/non-existent-id", nil)
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestBadRequestCases(t *testing.T) {
	_, _, mux, authService, cleanup := setupTestEnv(t)
	defer cleanup()

	sessionCookie := login(t, authService)

	t.Run("POST /api/notes with empty title and empty content returns 400", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title":"","content":"","template":""}`)
		req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("POST /api/notes with whitespace-only title and content returns 400", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title":"   ","content":"   ","template":""}`)
		req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("POST /api/notes with valid title succeeds", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title":"Valid Title","content":"","template":""}`)
		req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
}

func TestAuthRequired(t *testing.T) {
	_, _, mux, _, cleanup := setupTestEnv(t)
	defer cleanup()

	t.Run("GET /api/notes without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notes", nil)

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("POST /api/notes without auth returns 401", func(t *testing.T) {
		body := bytes.NewBufferString(`{"title":"Test","content":"Test"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}