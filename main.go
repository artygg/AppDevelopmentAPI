// main.go – финальная рабочая версия

package main

import (
    "AppDevelopmentAPI/websocket"
    "bytes"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/google/uuid"
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

/* ────────────────────────────  DATA  ──────────────────────────── */

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
    ID        string   `json:"id"`                // uuid
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

const quizTTL = 24 * time.Hour

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

func sendUpdate(u UpdateMessage) {
    if b, err := json.Marshal(u); err == nil {
        websocket.Broadcast <- b
    }
}

/* ─────────────────────────  CRUD HELPERS  ─────────────────────── */

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
    rows, _ := db.Query(`SELECT qid FROM mines WHERE place_id=$1 AND expires_at>now()`, pid)
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

/* ────────────────────────  HTTP HANDLERS  ─────────────────────── */

func placesHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        if places, err := getAllPlaces(db); err == nil {
            json.NewEncoder(w).Encode(places)
        } else {
            http.Error(w, "DB error", 500)
        }
    }
}

func iconLookupHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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
    }
}

func categoryIconsHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
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
    }
}

func createPlaceHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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
    }
}

func capturePlaceHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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
        if _, err := db.Exec(`UPDATE places SET captured=TRUE,user_captured=$1 WHERE id=$2`, req.User, req.PlaceID); err != nil {
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
    }
}

func mineHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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
    }
}

func finishHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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
        db.Exec(`INSERT INTO capture_attempts(place_id,user_name,correct,time_ms) VALUES($1,$2,$3,$4)`,
            f.PlaceID, f.User, f.Correct, f.TimeMs)

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
            f.PlaceID, f.Correct, f.TimeMs, f.User)

        var captured bool
        _ = db.QueryRow(`SELECT holder=$1 FROM place_scores WHERE place_id=$2`, f.User, f.PlaceID).Scan(&captured)
        json.NewEncoder(w).Encode(struct {
            Captured bool `json:"captured"`
        }{captured})
    }
}

func quizHandler(db *sql.DB, key string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
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
    }
}

/* ────────────────────  STUB IMAGE HANDLERS  ───────────────────── */

func UploadImageHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        http.Error(w, "not implemented", http.StatusNotImplemented)
    }
}

func GetImageByPlaceIDHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        http.Error(w, "not implemented", http.StatusNotImplemented)
    }
}

/* ─────────────────────────────  MAIN  ─────────────────────────── */

func main() {
    db := dbConnect()
    defer db.Close()
    key := os.Getenv("OPENAI_API_KEY")

    http.HandleFunc("/places", placesHandler(db))
    http.HandleFunc("/quiz", quizHandler(db, key))
    http.HandleFunc("/icon", iconLookupHandler(db))
    http.HandleFunc("/category_icons.json", categoryIconsHandler(db))

    http.HandleFunc("/api/places", createPlaceHandler(db))
    http.HandleFunc("/api/capture", capturePlaceHandler(db))
    http.HandleFunc("/api/mine", mineHandler(db))
    http.HandleFunc("/api/finish", finishHandler(db))

    http.HandleFunc("/upload-file", UploadImageHandler(db))
    http.HandleFunc("/get-image", GetImageByPlaceIDHandler(db))

    go websocket.HandleMessages()
    http.HandleFunc("/ws", websocket.WebSocketHandler)

    http.Handle("/", http.FileServer(http.Dir(".")))

    log.Println("API listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
