package database

import (
	"errors"

	"database/sql"
)

func GetNonce(wallet string) (string, error) {
	var err error
	var nonce sql.NullString

	row := db.QueryRow("SELECT get_nonce($1);", wallet)
	err = row.Scan(&nonce)
	if !nonce.Valid {
		return "", errors.New("no user")
	}

	return nonce.String, err
}

func Register(wallet string, nonce string) error {
	_, err := db.Exec("INSERT INTO users(wallet, nonce) VALUES ($1, $2);", wallet, nonce)

	return err
}

func SetNonce(wallet string, nonce string) error {
	_, err := db.Exec("UPDATE users SET nonce = $2 WHERE wallet = $1;", wallet, nonce)

	return err
}
