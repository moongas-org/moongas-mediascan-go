package main

import (
	"database/sql"
	"fmt"
	"log/slog"
)

func createTableArtist(db *sql.DB) error {
	sql := `CREATE TABLE artist (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"path" TEXT UNIQUE,
		"name" TEXT,
		"city" TEXT,
		"countrycode" TEXT,
		"regioncode" TEXT,
		"languagecode" TEXT,
		"hasartistyaml" BOOLEAN NOT NULL DEFAULT FALSE
	  );` // SQL Statement for Create Table

	slog.Info("Creating artist table")
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("create artist table: %w", err)
	}
	slog.Info("Created artist table")
	return nil
}

func insertArtist(db *sql.DB, path string, name string, city string, countrycode string, regioncode string, languagecode string, hasArtistYaml bool) error {
	slog.Debug("Inserting artist record", "path", path)
	_, err := db.Exec(`INSERT INTO artist(path, name, city, countrycode, regioncode, languagecode, hasartistyaml)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, path, name, city, countrycode, regioncode, languagecode, hasArtistYaml)
	if err != nil {
		return fmt.Errorf("insert artist %s: %w", path, err)
	}
	return nil
}

func displayArtists(db *sql.DB) error {
	row, err := db.Query("SELECT id, path, name, city, countrycode, regioncode, languagecode, hasartistyaml FROM artist ORDER BY name")
	if err != nil {
		return fmt.Errorf("query artists: %w", err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var path string
		var name string
		var city string
		var countrycode string
		var regioncode string
		var languagecode string
		var hasArtistYaml bool
		if err := row.Scan(&id, &path, &name, &city, &countrycode, &regioncode, &languagecode, &hasArtistYaml); err != nil {
			return fmt.Errorf("scan artist row: %w", err)
		}
		slog.Debug("Artist row", "id", id, "path", path, "name", name, "city", city, "country_code", countrycode, "region_code", regioncode, "language_code", languagecode, "has_artist_yaml", hasArtistYaml)
	}
	if err := row.Err(); err != nil {
		return fmt.Errorf("read artist rows: %w", err)
	}
	return nil
}
