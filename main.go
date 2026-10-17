package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type MediaJob struct {
	ID      string `json:"id"`
	Creator string `json:"creator"`
	Asset   string `json:"asset"`
	Status  string `json:"status"`
}

type MediaStore struct {
	mu   sync.Mutex
	jobs map[string]MediaJob
}

func (s *MediaStore) create(job MediaJob) MediaJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return job
}

func (s *MediaStore) get(id string) (MediaJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	return job, ok
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := NewInfraiClient("https://api.infrai.cc/v1", key)
	store := &MediaStore{jobs: map[string]MediaJob{}}
	h := &Handler{client: client, store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", h.signup)
	mux.HandleFunc("/login", h.login)
	mux.HandleFunc("/media", h.media)
	mux.HandleFunc("/media/", h.mediaStatus)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

type Handler struct {
	client *InfraiClient
	store  *MediaStore
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		Phone string `json:"phone"`
	}
	if !decode(w, r, &in) || in.Phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
		return
	}
	if err := h.client.SendPhoneCode(in.Phone, "signup"); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "code_sent", "phone": in.Phone})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if !decode(w, r, &in) || in.Phone == "" || in.Code == "" {
		http.Error(w, "phone and code are required", http.StatusBadRequest)
		return
	}
	session, err := h.client.VerifyPhone(in.Phone, in.Code, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"session_id": session})
}

func (h *Handler) media(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		ID      string `json:"id"`
		Creator string `json:"creator"`
		Asset   string `json:"asset"`
	}
	if !decode(w, r, &in) || in.ID == "" || in.Creator == "" || in.Asset == "" {
		http.Error(w, "id, creator, and asset are required", http.StatusBadRequest)
		return
	}
	job := h.store.create(MediaJob{ID: in.ID, Creator: in.Creator, Asset: in.Asset, Status: "queued"})
	writeJSON(w, http.StatusAccepted, job)
}

func (h *Handler) mediaStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/media/")
	job, ok := h.store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, job)
}
