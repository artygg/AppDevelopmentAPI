package main

import (
    "AppDevelopmentAPI/websocket"
    "bytes"
    "database/sql"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

type Place struct {
    ID           int      `json:"id"`
    Name         string   `json:"name"`
    Latitude     float64  `json:"latitude"`
    Longitude    float64  `json:"longitude"`
    CategoryID   int      `json:"category_id"`
    Captured     bool     `json:"captured"`
    UserCaptured *string  `json:"user_captured"`
}
type Question struct {
    Text    string   `json:"text"`
    Options []string `json:"options"`
    Answer  int      `json:"answer"`
}
type Quiz struct {
    PlaceID   int        `json:"place_id"`
    Questions []Question `json:"questions"`
}
type UpdateMessage struct {
    Status string `json:"status"`
    Time   string `json:"time"`
    Source string `json:"source"`
    PlaceID int   `json:"place_id,omitempty"`
    PlaceName string `json:"place_name,omitempty"`
}

func dbConnect() *sql.DB {
    _ = godotenv.Load()
    dbName := os.Getenv("DB_NAME")
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    return db
}

func sendUpdate(update UpdateMessage) {
    jsonMsg, err := json.Marshal(update)
    if err != nil {
        log.Println("JSON marshal error:", err)
        return
    }
    websocket.Broadcast <- jsonMsg
}

func getAllPlaces(db *sql.DB) ([]Place, error) {
    rows, err := db.Query("SELECT id, name, latitude, longitude, category_id, captured, user_captured FROM places")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var places []Place
    for rows.Next() {
        var p Place
        var userCaptured sql.NullString
        err := rows.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &userCaptured)
        if err != nil {
            return nil, err
        }
        if userCaptured.Valid {
            p.UserCaptured = &userCaptured.String
        }
        places = append(places, p)
    }
    return places, nil
}

func getPlaceByName(db *sql.DB, name string) (*Place, error) {
    row := db.QueryRow("SELECT id, name, latitude, longitude, category_id, captured, user_captured FROM places WHERE name = $1 LIMIT 1", name)
    var p Place
    var userCaptured sql.NullString
    err := row.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &userCaptured)
    if err != nil {
        return nil, err
    }
    if userCaptured.Valid {
        p.UserCaptured = &userCaptured.String
    }
    return &p, nil
}

func getPlaceByID(db *sql.DB, id int) (*Place, error) {
    row := db.QueryRow("SELECT id, name, latitude, longitude, category_id, captured, user_captured FROM places WHERE id = $1 LIMIT 1", id)
    var p Place
    var userCaptured sql.NullString
    err := row.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude, &p.CategoryID, &p.Captured, &userCaptured)
    if err != nil {
        return nil, err
    }
    if userCaptured.Valid {
        p.UserCaptured = &userCaptured.String
    }
    return &p, nil
}

func getQuizByPlaceID(db *sql.DB, placeID int) (*Quiz, error) {
    row := db.QueryRow("SELECT quiz_json FROM quizzes WHERE place_id = $1 LIMIT 1", placeID)
    var quizData []byte
    err := row.Scan(&quizData)
    if err != nil {
        return nil, err
    }
    var quiz Quiz
    err = json.Unmarshal(quizData, &quiz)
    if err != nil {
        return nil, err
    }
    return &quiz, nil
}

func storeQuizForPlace(db *sql.DB, placeID int, quiz Quiz) error {
    quizBytes, err := json.Marshal(quiz)
    if err != nil {
        return err
    }
    _, err = db.Exec("INSERT INTO quizzes (place_id, quiz_json) VALUES ($1, $2) ON CONFLICT (place_id) DO UPDATE SET quiz_json = $2", placeID, quizBytes)
    return err
}

// --- Web endpoints ---

func placesHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        places, err := getAllPlaces(db)
        if err != nil {
            http.Error(w, "No places found", http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(places)
    }
}

func quizHandler(db *sql.DB, openaiKey string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        placeName := r.URL.Query().Get("place")
        if placeName == "" {
            http.Error(w, "Missing place parameter", http.StatusBadRequest)
            return
        }
        place, err := getPlaceByName(db, placeName)
        if err != nil {
            http.Error(w, "Place not found", http.StatusNotFound)
            return
        }
        quiz, err := getQuizByPlaceID(db, place.ID)
        if err == nil && quiz != nil {
            json.NewEncoder(w).Encode(quiz)
            return
        }
        questions, err := generateQuizForPlace(place.Name, place.Latitude, place.Longitude, openaiKey)
        if err != nil {
            http.Error(w, "Failed to generate quiz", http.StatusInternalServerError)
            return
        }
        newQuiz := Quiz{PlaceID: place.ID, Questions: questions}
        _ = storeQuizForPlace(db, place.ID, newQuiz)
        json.NewEncoder(w).Encode(newQuiz)
    }
}

func createPlaceHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var newPlace Place
        if err := json.NewDecoder(r.Body).Decode(&newPlace); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }
        var id int
        err := db.QueryRow(
            "INSERT INTO places (name, latitude, longitude, category_id, captured, user_captured) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
            newPlace.Name, newPlace.Latitude, newPlace.Longitude, newPlace.CategoryID, false, nil,
        ).Scan(&id)
        if err != nil {
            http.Error(w, "Failed to insert place", http.StatusInternalServerError)
            return
        }
        update := UpdateMessage{
            Status: "added",
            Time:   time.Now().Format(time.RFC3339),
            Source: "Places",
            PlaceID: id,
            PlaceName: newPlace.Name,
        }
        sendUpdate(update)
        newPlace.ID = id
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(newPlace)
    }
}

// --- Capture/Update logic (via POST for marking places captured) ---
func capturePlaceHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var req struct {
            PlaceID int    `json:"place_id"`
            User    string `json:"user"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }
        // Update DB
        _, err := db.Exec("UPDATE places SET captured = TRUE, user_captured = $1 WHERE id = $2", req.User, req.PlaceID)
        if err != nil {
            http.Error(w, "Failed to update place", http.StatusInternalServerError)
            return
        }
        // Optionally fetch updated place
        updated, _ := getPlaceByID(db, req.PlaceID)
        // Send websocket update!
        update := UpdateMessage{
            Status: "captured",
            Time:   time.Now().Format(time.RFC3339),
            Source: "Capture",
            PlaceID: req.PlaceID,
            PlaceName: updated.Name,
        }
        sendUpdate(update)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(updated)
    }
}

// --- Quiz Generation ---
func generateQuizForPlace(name string, lat, lon float64, apiKey string) ([]Question, error) {
    prompt := fmt.Sprintf(`
    You are generating a quiz for a location-based mobile game.
    This quiz must be about the real-world place named "%s" located at latitude %.6f, longitude %.6f.
    There may be other places with the same name, but this quiz should be specific to the location at these coordinates.
    Use all information you can (such as city, country, nearby landmarks, or known facts) to make the first 3 questions about this place, or the city/region it is in.
    If information about the place itself is limited, use questions about the city or area where the place is located.
    Only if you cannot generate more, fill with general easy knowledge questions.
    Each question must be an object with:
      - a non-empty "text" field (the question in English)
      - an "options" field: 4 possible answers
      - an "answer" field: the 0-based index of the correct answer in "options"
    NEVER use the field "question", only "text".
    Do not use markdown or explanation, just respond with a JSON array.
    If any "text" value is empty, regenerate it with a real question in English.
    Return ONLY valid JSON, not markdown.
    `, name, lat, lon)
    requestBody := map[string]interface{}{
        "model": "gpt-3.5-turbo",
        "messages": []map[string]string{
            {"role": "system", "content": "You are a quiz generator for a location-based game."},
            {"role": "user", "content": prompt},
        },
        "temperature": 0.7,
        "max_tokens": 1000,
    }
    reqBytes, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBytes))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    var apiResp struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    if err := json.Unmarshal(respBody, &apiResp); err != nil {
        return nil, err
    }
    var questions []Question
    cleaned := apiResp.Choices[0].Message.Content
    cleanedBytes := []byte(cleaned)
    cleanedBytes = bytes.TrimPrefix(cleanedBytes, []byte("```json\n"))
    cleanedBytes = bytes.TrimSuffix(cleanedBytes, []byte("\n```"))
    if err := json.Unmarshal(cleanedBytes, &questions); err != nil {
        if err := json.Unmarshal([]byte(apiResp.Choices[0].Message.Content), &questions); err != nil {
            return nil, err
        }
    }
    return questions, nil
}

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
    db := dbConnect()
    defer db.Close()
    openaiKey := os.Getenv("OPENAI_API_KEY")

    http.HandleFunc("/places", placesHandler(db))
    http.HandleFunc("/quiz", quizHandler(db, openaiKey))
    http.HandleFunc("/api/places", createPlaceHandler(db))
    http.HandleFunc("/api/capture", capturePlaceHandler(db))

    http.Handle("/", http.FileServer(http.Dir(".")))

    go websocket.HandleMessages()
    http.HandleFunc("/ws", websocket.WebSocketHandler)

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
