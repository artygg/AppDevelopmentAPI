package models

import "github.com/jmoiron/sqlx"

func GetMineBalance(db *sqlx.DB, user string) (int, error) {
	var bal int
	err := db.Get(&bal,
		`SELECT balance
		   FROM user_mines
		  WHERE username = $1`, user)
	return bal, err
}

func GrantMine(db *sqlx.DB, user string, n int) error {
	_, err := db.Exec(`
		INSERT INTO user_mines (username, balance)
		VALUES ($1,$2)
		ON CONFLICT (username) DO
		    UPDATE SET balance = user_mines.balance + $2`,
		user, n)
	return err
}

func ConsumeMine(db *sqlx.DB, user string, n int) error {
	_, err := db.Exec(`
		UPDATE user_mines
		   SET balance = balance - $2
		 WHERE username = $1
		   AND balance  >= $2`, user, n)
	return err
}
