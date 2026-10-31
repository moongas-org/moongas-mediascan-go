package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func rowIDByPath(db *sql.DB, table, path string) (sql.NullInt64, error) {
	var query string
	switch table {
	case "artist":
		query = "SELECT id FROM artist WHERE path = ?"
	case "album":
		query = "SELECT id FROM album WHERE path = ?"
	case "track":
		query = "SELECT id FROM track WHERE path = ?"
	default:
		return sql.NullInt64{}, fmt.Errorf("unsupported relationship table %q", table)
	}
	var id sql.NullInt64
	if err := db.QueryRow(query, path).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("WARNING: no %s row for %s; storing NULL %sid", table, path, table)
			return sql.NullInt64{}, nil
		}
		return sql.NullInt64{}, fmt.Errorf("find %s ID for %s: %w", table, path, err)
	}
	return id, nil
}
