package models

import (
	"BitGainz-Bittrex/lib/utils"
	"os"
	"time"

	// for mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// ConnectSQL function
func ConnectSQL(duration int) *sqlx.DB {
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	database := os.Getenv("DB_DATABASE")

	// Connect DB from environment
	db, _ := sqlx.Open("mysql", user+":"+password+"@tcp("+host+")/"+database)

	err := db.Ping()
	utils.HandleError(err)

	db.SetConnMaxLifetime(time.Second * time.Duration(duration+30))
	db.SetMaxIdleConns(0)
	return db
}
