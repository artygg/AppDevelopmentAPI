package models

import (
    "database/sql"
    "encoding/json"
    "time"
)

type Quiz struct {
    PlaceID   int        `json:"place_id"`
    Questions []Question `json:"questions"`
    UpdatedAt time.Time  `json:"-"`
}

func GetByPlaceID(db *sql.DB, pid int) (*Quiz, error) {
    var raw []byte
    var upd time.Time
    if err := db.QueryRow(
        `SELECT quiz_json,updated_at FROM quizzes WHERE place_id=$1`, pid,
    ).Scan(&raw, &upd); err != nil {
        return nil, err
    }
    var q Quiz
    if err := json.Unmarshal(raw, &q); err != nil {
        return nil, err
    }
    q.UpdatedAt = upd
    return &q, nil
}

func Store(db *sql.DB, pid int, q Quiz) error {
    b, _ := json.Marshal(q)
    _, _ = db.Exec(`DELETE FROM quizzes WHERE place_id=$1`, pid)
    _, err := db.Exec(
        `INSERT INTO quizzes(place_id,quiz_json,updated_at)
         VALUES($1,$2,now())`, pid, b,
    )
    return err
}

