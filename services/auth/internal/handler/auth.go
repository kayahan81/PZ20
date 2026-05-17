package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"tech-ip-sem2/shared/logger"
	"tech-ip-sem2/shared/middleware"

	"go.uber.org/zap"
)

const validToken = "demo-token"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AuthHandler struct {
	log *zap.Logger
}

func NewAuthHandler(log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		log: log,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := logger.WithRequestID(h.log, requestID)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("invalid login request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Info("login attempt", zap.String("username", req.Username))

	resp := LoginResponse{
		AccessToken: validToken,
		TokenType:   "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.Info("login successful", zap.String("username", req.Username))
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log := logger.WithRequestID(h.log, requestID)

	authHeader := r.Header.Get("Authorization")

	resp := VerifyResponse{}
	status := http.StatusOK

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		resp.Valid = false
		resp.Error = "unauthorized"
		status = http.StatusUnauthorized
		log.Warn("missing or invalid auth header")
	} else {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == validToken {
			resp.Valid = true
			resp.Subject = "student"
			log.Info("token verified", zap.String("subject", resp.Subject))
		} else {
			resp.Valid = false
			resp.Error = "unauthorized"
			status = http.StatusUnauthorized
			log.Warn("invalid token")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
