package db

import (
    "database/sql"
    "os"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
)

func Open() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(30 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
