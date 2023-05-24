package database

import (
	"fmt"

	_ "github.com/lib/pq"

	"database/sql"

	. "backend/config"
)

var (
	db *sql.DB
)

func Init() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		Config.Database.Host, Config.Database.Port, Config.Database.Username, Config.Database.Password, Config.Database.Name)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	if err := redisInit(); err != nil {
		return err
	}

	return nil
}
