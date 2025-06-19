package models

import "database/sql"

func IncCaptured(db *sql.DB, user string) error {
    _, err := db.Exec(`
        INSERT INTO player_totals(user_name, captured_count)
        VALUES ($1, 1)
        ON CONFLICT (user_name) DO UPDATE
          SET captured_count = player_totals.captured_count + 1,
              updated_at     = now()`, user)
    return err
}

type Leader struct {
    User  string `json:"user"`
    Score int    `json:"captured"`
    Rank  int    `json:"rank"`
}

func Leaderboard(db *sql.DB, limit int) ([]Leader, error) {
    rows, err := db.Query(`
        SELECT user_name,
               captured_count,
               ROW_NUMBER() OVER (ORDER BY captured_count DESC) AS r
        FROM   player_totals
        ORDER  BY captured_count DESC
        LIMIT  $1`, limit)
    if err != nil { return nil, err }
    defer rows.Close()

    var list []Leader
    for rows.Next() {
        var l Leader
        if err := rows.Scan(&l.User, &l.Score, &l.Rank); err != nil {
            return nil, err
        }
        list = append(list, l)
    }
    return list, nil
}
