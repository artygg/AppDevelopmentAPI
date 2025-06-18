package models

import (
	"database/sql"
	"time"
)

type Place struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	CategoryID   int        `json:"category_id"`
	Captured     bool       `json:"captured"`
	UserCaptured *string    `json:"user_captured,omitempty"`
	Cooldown     *time.Time `json:"cooldown_until,omitempty"`
}

func GetAll(db *sql.DB, user string) ([]Place, error) {
	rows, err := db.Query(`
	    SELECT p.id,p.name,p.latitude,p.longitude,p.category_id,
	           p.captured,p.user_captured,c.cooldown_until
	    FROM places p
	    LEFT JOIN place_cooldowns c
	           ON c.place_id=p.id AND c.user_name=$1`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Place
	for rows.Next() {
		var p Place
		var uc sql.NullString
		var cd sql.NullTime
		_ = rows.Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude,
			&p.CategoryID, &p.Captured, &uc, &cd)
		if uc.Valid {
			p.UserCaptured = &uc.String
		}
		if cd.Valid {
			t := cd.Time
			p.Cooldown = &t
		}
		list = append(list, p)
	}
	return list, nil
}

func GetByID(db *sql.DB, id int) (*Place, error) {
	var p Place
	var uc sql.NullString
	err := db.QueryRow(`SELECT id,name,latitude,longitude,category_id,
	                     captured,user_captured FROM places WHERE id=$1`, id).
		Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude,
			&p.CategoryID, &p.Captured, &uc)
	if err != nil {
		return nil, err
	}
	if uc.Valid {
		p.UserCaptured = &uc.String
	}
	return &p, nil
}

func GetByName(db *sql.DB, name string) (*Place, error) {
    var p Place
    var uc sql.NullString
    err := db.QueryRow(`
        SELECT id,name,latitude,longitude,category_id,
               captured,user_captured
        FROM places
        WHERE name = $1 LIMIT 1`, name).
        Scan(&p.ID, &p.Name, &p.Latitude, &p.Longitude,
            &p.CategoryID, &p.Captured, &uc)
    if err != nil {
        return nil, err
    }
    if uc.Valid {
        p.UserCaptured = &uc.String
    }
    return &p, nil
}
