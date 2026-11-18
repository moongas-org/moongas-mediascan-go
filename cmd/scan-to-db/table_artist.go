package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
)

func createTableArtist(db *sql.DB) {
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
	statement, err := db.Prepare(sql) // Prepare SQL Statement
	if err != nil {
		slog.Error("Unable to prepare artist table", "error", err)
		os.Exit(1)
	}
	_, err = statement.Exec()
	if err != nil {
		slog.Error("Unable to create artist table", "error", err)
		os.Exit(1)
	}
	slog.Info("Created artist table")
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

func displayArtists(db *sql.DB) {
	row, err := db.Query("SELECT id, path, name, city, countrycode, regioncode, languagecode, hasartistyaml FROM artist ORDER BY name")
	if err != nil {
		slog.Error("Unable to query artists", "error", err)
		os.Exit(1)
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
			slog.Error("Unable to scan artist row", "error", err)
			os.Exit(1)
		}
		slog.Debug("Artist row", "id", id, "path", path, "name", name, "city", city, "country_code", countrycode, "region_code", regioncode, "language_code", languagecode, "has_artist_yaml", hasArtistYaml)
	}
	if err := row.Err(); err != nil {
		slog.Error("Unable to read artist rows", "error", err)
		os.Exit(1)
	}
}
