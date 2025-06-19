// main.go – Enhanced version with secure authentication

package main

import (
	_ "AppDevelopmentAPI/internal/db"
	"AppDevelopmentAPI/internal/handlers"
	"AppDevelopmentAPI/websocket"
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

/* ────────────────────────────  DATA STRUCTURES  ──────────────────────────── */

type Place struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	CategoryID   int     `json:"category_id"`
	Captured     bool    `json:"captured"`
	UserCaptured *string `json:"user_captured"`
}

type Question struct {
	ID        string   `json:"id"` // uuid
	Text      string   `json:"text"`
	Options   []string `json:"options"`
	Answer    int      `json:"answer"`
	TimeLimit *int     `json:"timeLimit,omitempty"`
}

type Quiz struct {
	PlaceID   int        `json:"place_id"`
	Questions []Question `json:"questions"`
	UpdatedAt time.Time  `json:"-"`
}

type UpdateMessage struct {
	Status    string `json:"status"`
	Time      string `json:"time"`
	Source    string `json:"source"`
	PlaceID   int    `json:"place_id,omitempty"`
	PlaceName string `json:"place_name,omitempty"`
}

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Password     string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	FailedLogins int        `json:"-"`
	LockedUntil  *time.Time `json:"-"`
	MapImageUrl  *string    `json:"map_image_url"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User         *User  `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenData struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IsRevoked bool      `json:"is_revoked"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

/* ────────────────────────────  SECURITY CONSTANTS  ─────────────────────────── */

const (
	quizTTL                = 24 * time.Hour
	maxFailedLogins        = 5
	accountLockDuration    = 30 * time.Minute
	jwtExpiration          = 15 * time.Minute
	refreshTokenExpiration = 7 * 24 * time.Hour
	bcryptCost             = 12
	minPasswordLength      = 8
	maxPasswordLength      = 128
	minUsernameLength      = 3
	maxUsernameLength      = 30
)

/* ────────────────────────────  GLOBAL VARIABLES  ───────────────────────────── */

var (
	jwtSecret        []byte
	rateLimiters     = make(map[string]*rate.Limiter)
	rateLimiterMutex sync.RWMutex
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex    = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

/* ────────────────────────  RATE LIMITING  ────────────────────── */

func getRateLimiter(ip string) *rate.Limiter {
	rateLimiterMutex.RLock()
	limiter, exists := rateLimiters[ip]
	rateLimiterMutex.RUnlock()

	if !exists {
		rateLimiterMutex.Lock()
		limiter, exists = rateLimiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(time.Minute), 100)
			rateLimiters[ip] = limiter
		}
		rateLimiterMutex.Unlock()
	}

	return limiter
}

func rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := getRateLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.Split(xff, ",")[0]
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

/* ────────────────────────  VALIDATION  ─────────────────────────── */

func validateEmail(email string) error {
	if len(email) > 254 {
		return fmt.Errorf("email too long")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateUsername(username string) error {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return fmt.Errorf("username must be between %d and %d characters", minUsernameLength, maxUsernameLength)
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, hyphens, and underscores")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return fmt.Errorf("password must be between %d and %d characters", minPasswordLength, maxPasswordLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character")
	}

	return nil
}

/* ────────────────────────  SECURITY UTILITIES  ─────────────────────────── */

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWT(userID int, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

/* ────────────────────────  DB & UTILITIES  ────────────────────── */

func dbConnect() *sql.DB {
	_ = godotenv.Load()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initDB(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(30) UNIQUE NOT NULL,
			email VARCHAR(254) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			is_active BOOLEAN DEFAULT TRUE,
			last_login_at TIMESTAMP,
			failed_logins INTEGER DEFAULT 0,
			locked_until TIMESTAMP,
		    map_image_url TEXT
		)
	`)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id VARCHAR(36) PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			is_revoked BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		log.Fatal("Failed to create refresh_tokens table:", err)
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`)
}

func sendUpdate(u UpdateMessage) {
	if b, err := json.Marshal(u); err == nil {
		websocket.Broadcast <- b
	}
}

/* ─────────────────────────  AUTH CRUD HELPERS  ─────────────────────── */

func getUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime
	var mapImageURL sql.NullString

	err := db.QueryRow(`
		SELECT id, username, email, password_hash, created_at, updated_at, 
		       is_active, last_login_at, failed_logins, locked_until, map_image_url
		FROM users WHERE username = $1`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive,
		&lastLoginAt, &user.FailedLogins, &lockedUntil, &mapImageURL)

	if err != nil {
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if mapImageURL.Valid {
		user.MapImageUrl = &mapImageURL.String
	}

	return user, nil
}

func getUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{}
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime
	var mapImageURL sql.NullString

	err := db.QueryRow(`
		SELECT id, username, email, password_hash, created_at, updated_at, 
		       is_active, last_login_at, failed_logins, locked_until, map_image_url
		FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive,
		&lastLoginAt, &user.FailedLogins, &lockedUntil, &mapImageURL)

	if err != nil {
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if mapImageURL.Valid {
		user.MapImageUrl = &mapImageURL.String
	}

	return user, nil
}

func createUser(db *sql.DB, username, email, password string) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = db.QueryRow(`
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, created_at, updated_at, is_active`,
		username, email, hashedPassword).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive)

	return user, err
}

func updateLoginAttempt(db *sql.DB, userID int, success bool) error {
	if success {
		_, err := db.Exec(`
			UPDATE users 
			SET last_login_at = NOW(), failed_logins = 0, locked_until = NULL, updated_at = NOW()
			WHERE id = $1`, userID)
		return err
	} else {
		_, err := db.Exec(`
			UPDATE users 
			SET failed_logins = failed_logins + 1,
			    locked_until = CASE 
			        WHEN failed_logins + 1 >= $1 THEN NOW() + INTERVAL '%d minutes'
			        ELSE locked_until
			    END,
			    updated_at = NOW()
			WHERE id = $2`, maxFailedLogins, int(accountLockDuration.Minutes()), userID)
		return err
	}
}

func createRefreshToken(db *sql.DB, userID int) (string, error) {
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	tokenHash, err := hashPassword(token)
	if err != nil {
		return "", err
	}

	tokenID := uuid.NewString()
	_, err = db.Exec(`
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`,
		tokenID, userID, tokenHash, time.Now().Add(refreshTokenExpiration))

	if err != nil {
		return "", err
	}

	return tokenID + ":" + token, nil
}

func validateRefreshToken(db *sql.DB, tokenString string) (*RefreshTokenData, error) {
	parts := strings.Split(tokenString, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	tokenID, token := parts[0], parts[1]

	var tokenData RefreshTokenData
	err := db.QueryRow(`
		SELECT id, user_id, token_hash, expires_at, created_at, is_revoked
		FROM refresh_tokens
		WHERE id = $1 AND expires_at > NOW() AND is_revoked = FALSE`,
		tokenID).Scan(
		&tokenData.ID, &tokenData.UserID, &tokenData.Token,
		&tokenData.ExpiresAt, &tokenData.CreatedAt, &tokenData.IsRevoked)

	if err != nil {
		return nil, err
	}

	if !checkPasswordHash(token, tokenData.Token) {
		return nil, fmt.Errorf("invalid token")
	}

	return &tokenData, nil
}

func revokeRefreshToken(db *sql.DB, tokenID string) error {
	_, err := db.Exec(`
		UPDATE refresh_tokens 
		SET is_revoked = TRUE 
		WHERE id = $1`, tokenID)
	return err
}

func cleanupExpiredTokens(db *sql.DB) {
	_, err := db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		log.Printf("Error cleaning up expired tokens: %v", err)
	}
}

/* ────────────────────────  PROFILE IMAGE HANDLERS  ─────────────────────────── */

type ProfileImage struct {
	ID       int    `json:"id"`
	ImageURL string `json:"image_url"`
}

type UpdateMapImageRequest struct {
	ImageURL string `json:"image_url"`
}

func getAllProfileImagesHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`SELECT id, image_url FROM profile_images ORDER BY id`)
		if err != nil {
			log.Printf("Error fetching profile images: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var images []ProfileImage
		for rows.Next() {
			var img ProfileImage
			if err := rows.Scan(&img.ID, &img.ImageURL); err != nil {
				log.Printf("Error scanning profile image: %v", err)
				continue
			}
			images = append(images, img)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating profile images: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(images)
	})
}

func updateUserMapImageHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, err := strconv.Atoi(r.Header.Get("X-User-ID"))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var req UpdateMapImageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(req.ImageURL) == "" {
			http.Error(w, "Image URL cannot be empty", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`
			UPDATE users 
			SET map_image_url = $1, updated_at = NOW() 
			WHERE id = $2`, req.ImageURL, userID)

		if err != nil {
			log.Printf("Error updating user map image: %v", err)
			http.Error(w, "Failed to update map image", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"message":   "Map image updated successfully",
			"image_url": req.ImageURL,
		})
	}))
}

/* ─────────────────────────  PLACE CRUD HELPERS  ─────────────────────── */

func getAllPlaces(db *sql.DB) ([]Place, error) {
	rows, err := db.Query(`SELECT id,name,latitude,longitude,category_id,captured,user_captured FROM places`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Place
	for rows.Next() {
		var p Place
		var uc sql.NullString
		_ = rows.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &uc)
		if uc.Valid {
			p.UserCaptured = &uc.String
		}
		list = append(list, p)
	}
	return list, nil
}

func getPlaceByID(db *sql.DB, id int) (*Place, error) {
	row := db.QueryRow(`SELECT id,name,latitude,longitude,category_id,captured,user_captured FROM places WHERE id=$1`, id)
	var p Place
	var uc sql.NullString
	if err := row.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &uc); err != nil {
		return nil, err
	}
	if uc.Valid {
		p.UserCaptured = &uc.String
	}
	return &p, nil
}

func getPlaceByName(db *sql.DB, name string) (*Place, error) {
	row := db.QueryRow(`SELECT id,name,latitude,longitude,category_id,captured,user_captured FROM places WHERE name=$1 LIMIT 1`, name)
	var p Place
	var uc sql.NullString
	if err := row.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &uc); err != nil {
		return nil, err
	}
	if uc.Valid {
		p.UserCaptured = &uc.String
	}
	return &p, nil
}

func getQuizByPlaceID(db *sql.DB, pid int) (*Quiz, error) {
	row := db.QueryRow(`SELECT quiz_json,updated_at FROM quizzes WHERE place_id=$1`, pid)
	var raw []byte
	var upd time.Time
	if err := row.Scan(&raw, &upd); err != nil {
		return nil, err
	}
	var q Quiz
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, err
	}
	q.UpdatedAt = upd
	return &q, nil
}

func storeQuizForPlace(db *sql.DB, pid int, q Quiz) error {
	b, _ := json.Marshal(q)
	_, _ = db.Exec(`DELETE FROM quizzes WHERE place_id=$1`, pid)
	_, err := db.Exec(`INSERT INTO quizzes(place_id,quiz_json,updated_at) VALUES($1,$2,now())`, pid, b)
	return err
}

/* ────────────────────────────  MINES  ─────────────────────────── */

func applyMines(db *sql.DB, pid int, qs []Question) {
	if db == nil {
		log.Println("applyMines: db is nil")
		return
	}
	rows, err := db.Query(`SELECT qid FROM mines WHERE place_id=$1 AND expires_at>now()`, pid)
	if err != nil {
		log.Println("applyMines: db.Query error:", err)
		return
	}
	defer rows.Close()
	mined := map[string]struct{}{}
	for rows.Next() {
		var id string
		_ = rows.Scan(&id)
		mined[id] = struct{}{}
	}
	for i := range qs {
		if _, ok := mined[qs[i].ID]; ok {
			t := 5
			qs[i].TimeLimit = &t
		}
	}
}

/* ─────────────────────  AI QUIZ GENERATOR  ────────────────────── */

func generateQuizForPlace(name string, lat, lon float64, key string) ([]Question, error) {
	prompt := fmt.Sprintf(
		`Generate 7 JSON questions (array) about "%s" (%.5f,%.5f). Each item must be {"text":...,"options":[4],"answer":number}. No markdown.`,
		name, lat, lon)

	body := map[string]any{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a quiz generator for a location-based game."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  1000,
	}
	reqBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBytes))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var api struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return nil, err
	}

	var qs []Question
	_ = json.Unmarshal([]byte(api.Choices[0].Message.Content), &qs)
	for i := range qs {
		qs[i].ID = uuid.NewString()
	}
	return qs, nil
}

/* ────────────────────────  MIDDLEWARE  ─────────────────────────── */

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := validateJWT(bearerToken[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		r.Header.Set("X-Username", claims.Username)

		next(w, r)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

/* ────────────────────────  AUTH HANDLERS  ─────────────────────────── */

func registerHandler(db *sql.DB) http.HandlerFunc {
	return rateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := validateUsername(req.Username); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateEmail(req.Email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validatePassword(req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := getUserByUsername(db, req.Username); err == nil {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}

		if _, err := getUserByEmail(db, req.Email); err == nil {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}

		user, err := createUser(db, req.Username, req.Email, req.Password)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			log.Printf("Error generating JWT: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := createRefreshToken(db, user.ID)
		if err != nil {
			log.Printf("Error creating refresh token: %v", err)
			http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AuthResponse{
			User:         user,
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(jwtExpiration).Unix(),
		})
	})
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return rateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		user, err := getUserByUsername(db, req.Username)
		if err != nil {
			if emailRegex.MatchString(req.Username) {
				user, err = getUserByEmail(db, req.Username)
			}
			if err != nil {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
		}

		if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
			http.Error(w, "Account is temporarily locked", http.StatusLocked)
			return
		}

		if !user.IsActive {
			http.Error(w, "Account is disabled", http.StatusUnauthorized)
			return
		}

		if !checkPasswordHash(req.Password, user.Password) {
			updateLoginAttempt(db, user.ID, false)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := updateLoginAttempt(db, user.ID, true); err != nil {
			log.Printf("Error updating login attempt: %v", err)
		}

		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			log.Printf("Error generating JWT: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := createRefreshToken(db, user.ID)
		if err != nil {
			log.Printf("Error creating refresh token: %v", err)
			http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
			return
		}

		user.Password = ""

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			User:         user,
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(jwtExpiration).Unix(),
		})
	})
}

func getUserByID(db *sql.DB, id int) (*User, error) {
	user := &User{}
	var lastLoginAt sql.NullTime
	var lockedUntil sql.NullTime
	var mapImageURL sql.NullString

	err := db.QueryRow(`
		SELECT id, username, email, password_hash, created_at, updated_at, 
		       is_active, last_login_at, failed_logins, locked_until, map_image_url
		FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt, &user.IsActive,
		&lastLoginAt, &user.FailedLogins, &lockedUntil, &mapImageURL)

	if err != nil {
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}
	if mapImageURL.Valid {
		user.MapImageUrl = &mapImageURL.String
	}

	return user, nil
}

func refreshTokenHandler(db *sql.DB) http.HandlerFunc {
	return rateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		tokenData, err := validateRefreshToken(db, req.RefreshToken)
		if err != nil {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		user, err := getUserByID(db, tokenData.UserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		if !user.IsActive {
			http.Error(w, "Account is disabled", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(req.RefreshToken, ":")
		if len(parts) == 2 {
			revokeRefreshToken(db, parts[0])
		}

		token, err := generateJWT(user.ID, user.Username)
		if err != nil {
			log.Printf("Error generating JWT: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := createRefreshToken(db, user.ID)
		if err != nil {
			log.Printf("Error creating refresh token: %v", err)
			http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AuthResponse{
			User:         user,
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(jwtExpiration).Unix(),
		})
	})
}

func logoutHandler(db *sql.DB) http.HandlerFunc {
	return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
			return
		}

		parts := strings.Split(req.RefreshToken, ":")
		if len(parts) == 2 {
			revokeRefreshToken(db, parts[0])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
	})
}

func profileHandler(db *sql.DB) http.HandlerFunc {
	return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.Atoi(r.Header.Get("X-User-ID"))

		user := &User{}
		var lastLoginAt sql.NullTime
		var mapImageURL sql.NullString

		err := db.QueryRow(`
			SELECT id, username, email, created_at, updated_at, is_active, last_login_at, map_image_url
			FROM users WHERE id = $1`, userID).Scan(
			&user.ID, &user.Username, &user.Email,
			&user.CreatedAt, &user.UpdatedAt, &user.IsActive,
			&lastLoginAt, &mapImageURL,
		)

		if mapImageURL.Valid {
			user.MapImageUrl = &mapImageURL.String
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})
}

/* ────────────────────────  HTTP HANDLERS  ─────────────────────── */

func placesHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if places, err := getAllPlaces(db); err == nil {
			json.NewEncoder(w).Encode(places)
		} else {
			http.Error(w, "DB error", 500)
		}
	})
}

func iconLookupHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		cid := r.URL.Query().Get("category_id")
		if cid == "" {
			http.Error(w, "missing id", 400)
			return
		}
		var name string
		if db.QueryRow(`SELECT icon_name FROM category_icons WHERE category_id=$1`, cid).Scan(&name) != nil {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(name)
	})
}

func categoryIconsHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(func(w http.ResponseWriter, _ *http.Request) {
		rows, _ := db.Query(`SELECT category_id,icon_name FROM category_icons`)
		defer rows.Close()
		m := map[string]string{}
		for rows.Next() {
			var id int
			var n string
			_ = rows.Scan(&id, &n)
			m[fmt.Sprint(id)] = n
		}
		json.NewEncoder(w).Encode(m)
	})
}

func createPlaceHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", 405)
			return
		}
		var p Place
		if json.NewDecoder(r.Body).Decode(&p) != nil {
			http.Error(w, "bad body", 400)
			return
		}
		if db.QueryRow(`INSERT INTO places(name,latitude,longitude,category_id,captured) VALUES($1,$2,$3,$4,FALSE) RETURNING id`,
			p.Name, p.Latitude, p.Longitude, p.CategoryID).Scan(&p.ID) != nil {
			http.Error(w, "DB error", 500)
			return
		}
		sendUpdate(UpdateMessage{"added", time.Now().Format(time.RFC3339), "Places", p.ID, p.Name})
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(p)
	}))
}

func capturePlaceHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", 405)
			return
		}
		var req struct {
			PlaceID int    `json:"place_id"`
			User    string `json:"user"`
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			http.Error(w, "bad body", 400)
			return
		}

		username := r.Header.Get("X-Username")
		if username == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		if _, err := db.Exec(`UPDATE places SET captured=TRUE,user_captured=$1 WHERE id=$2`, username, req.PlaceID); err != nil {
			http.Error(w, "DB error", 500)
			return
		}
		p, _ := getPlaceByID(db, req.PlaceID)

		qs, _ := generateQuizForPlace(p.Name, p.Latitude, p.Longitude, os.Getenv("OPENAI_API_KEY"))
		newQ := Quiz{PlaceID: p.ID, Questions: qs}
		_ = storeQuizForPlace(db, p.ID, newQ)

		sendUpdate(UpdateMessage{"captured", time.Now().Format(time.RFC3339), "Capture", p.ID, p.Name})
		json.NewEncoder(w).Encode(struct {
			Place *Place `json:"place"`
			Quiz  *Quiz  `json:"quiz"`
		}{p, &newQ})
	}))
}

func mineHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", 405)
			return
		}
		var m struct {
			PlaceID int    `json:"place_id"`
			QID     string `json:"qid"`
		}
		if json.NewDecoder(r.Body).Decode(&m) != nil {
			http.Error(w, "bad body", 400)
			return
		}
		if _, err := db.Exec(`
          INSERT INTO mines(place_id,qid,expires_at)
          VALUES($1,$2,now()+interval '24 hours')
          ON CONFLICT(place_id,qid) DO UPDATE
          SET expires_at=now()+interval '24 hours'`, m.PlaceID, m.QID); err != nil {
			http.Error(w, "DB error", 500)
			return
		}
		w.WriteHeader(201)
	}))
}

func finishHandler(db *sql.DB) http.HandlerFunc {
	return corsMiddleware(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", 405)
			return
		}
		var f struct {
			PlaceID int    `json:"place_id"`
			User    string `json:"user"`
			Correct int    `json:"correct"`
			TimeMs  int64  `json:"elapsed_ms"`
		}
		if json.NewDecoder(r.Body).Decode(&f) != nil {
			http.Error(w, "bad body", 400)
			return
		}

		username := r.Header.Get("X-Username")
		if username == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		db.Exec(`INSERT INTO capture_attempts(place_id,user_name,correct,time_ms) VALUES($1,$2,$3,$4)`,
			f.PlaceID, username, f.Correct, f.TimeMs)

		db.Exec(`
          INSERT INTO place_scores(place_id,best_correct,best_time_ms,holder)
          VALUES($1,$2,$3,$4)
          ON CONFLICT (place_id) DO UPDATE
          SET best_correct = EXCLUDED.best_correct,
              best_time_ms = EXCLUDED.best_time_ms,
              holder       = EXCLUDED.holder,
              updated_at   = now()
          WHERE EXCLUDED.best_correct > place_scores.best_correct
             OR (EXCLUDED.best_correct = place_scores.best_correct
                 AND EXCLUDED.best_time_ms < place_scores.best_time_ms)`,
			f.PlaceID, f.Correct, f.TimeMs, username)

		var captured bool
		_ = db.QueryRow(`SELECT holder=$1 FROM place_scores WHERE place_id=$2`, username, f.PlaceID).Scan(&captured)
		json.NewEncoder(w).Encode(struct {
			Captured bool `json:"captured"`
		}{captured})
	}))
}

func quizHandler(db *sql.DB, key string) http.HandlerFunc {
	return corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var place *Place
		var err error
		if id := r.URL.Query().Get("place_id"); id != "" {
			var pid int
			fmt.Sscanf(id, "%d", &pid)
			place, err = getPlaceByID(db, pid)
		} else if n := r.URL.Query().Get("place"); n != "" {
			place, err = getPlaceByName(db, n)
		}
		if err != nil || place == nil {
			http.Error(w, "not found", 404)
			return
		}

		if q, _ := getQuizByPlaceID(db, place.ID); q != nil && time.Since(q.UpdatedAt) < quizTTL {
			json.NewEncoder(w).Encode(q)
			return
		}

		qs, _ := generateQuizForPlace(place.Name, place.Latitude, place.Longitude, key)
		applyMines(db, place.ID, qs)

		newQ := Quiz{PlaceID: place.ID, Questions: qs}
		_ = storeQuizForPlace(db, place.ID, newQ)
		json.NewEncoder(w).Encode(newQ)
	})
}

func capturedPlacesHandler(db *sql.DB) http.HandlerFunc {
	return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("X-Username")
		if username == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		rows, err := db.Query(`SELECT id, name, latitude, longitude, category_id, captured, user_captured FROM places WHERE user_captured = $1`, username)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var places []Place
		for rows.Next() {
			var p Place
			var uc sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &uc); err != nil {
				continue
			}
			if uc.Valid {
				p.UserCaptured = &uc.String
			}
			places = append(places, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(places)
	})
}

/* ─────────────────────────────  MAIN  ─────────────────────────── */

func main() {
	_ = godotenv.Load()

	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	jwtSecret = []byte(jwtSecretStr)

	db := dbConnect()
	if db == nil {
		log.Fatal("Database connection failed")
	}
	defer db.Close()

	initDB(db)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanupExpiredTokens(db)
			}
		}
	}()

	//key := os.Getenv("OPENAI_API_KEY")

	h := handlers.New(db, os.Getenv("OPENAI_API_KEY"))

	// Public routes
	//http.HandleFunc("/places", placesHandler(db))
	//http.HandleFunc("/quiz", quizHandler(db, key))
	//http.HandleFunc("/icon", iconLookupHandler(db))
	//http.HandleFunc("/category_icons.json", categoryIconsHandler(db))

	// Protected routes
	//http.HandleFunc("/api/places", createPlaceHandler(db))
	//http.HandleFunc("/api/capture", capturePlaceHandler(db))
	//http.HandleFunc("/api/mine", mineHandler(db))
	//http.HandleFunc("/api/finish", finishHandler(db))
	//http.HandleFunc("/api/captured_places", capturedPlacesHandler(db))

	mux := http.NewServeMux()
	// Auth routes
	mux.HandleFunc("/auth/register", registerHandler(db))
	mux.HandleFunc("/auth/login", loginHandler(db))
	mux.HandleFunc("/auth/refresh", refreshTokenHandler(db))
	mux.HandleFunc("/auth/logout", logoutHandler(db))
	mux.HandleFunc("/auth/profile", profileHandler(db))

	// Public routes
	mux.HandleFunc("/places", corsMiddleware(h.Places))
	mux.HandleFunc("/quiz", corsMiddleware(h.Quiz))
	mux.HandleFunc("/icon", corsMiddleware(h.IconLookup))
	mux.HandleFunc("/category_icons.json", corsMiddleware(h.CategoryIcons))
	mux.HandleFunc("/leaderboard", h.Leaderboard)

	// Protected routes
	mux.HandleFunc("/api/places", corsMiddleware(authMiddleware(h.CreatePlace)))
	mux.HandleFunc("/api/capture", corsMiddleware(authMiddleware(h.Capture)))
	mux.HandleFunc("/api/mine", corsMiddleware(authMiddleware(h.Mine)))
	mux.HandleFunc("/api/finish", corsMiddleware(authMiddleware(h.Finish)))
	mux.HandleFunc("/api/captured_places", capturedPlacesHandler(db))

	// Image routes
	mux.HandleFunc("/upload-file", UploadImageHandler(db))
	mux.HandleFunc("/get-image", GetImageByPlaceIDHandler(db))

	// WebSocket
	go websocket.HandleMessages()
	mux.HandleFunc("/ws", websocket.WebSocketHandler)

	// Static files
	mux.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("API listening on :8080")
	log.Println("Authentication endpoints available:")
	log.Println("  POST /auth/register - Register new user")
	log.Println("  POST /auth/login - Login user")
	log.Println("  POST /auth/refresh - Refresh JWT token")
	log.Println("  POST /auth/logout - Logout user")
	log.Println("  GET /auth/profile - Get user profile (requires auth)")
	log.Println()
	log.Println("Protected endpoints require 'Authorization: Bearer <token>' header")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
