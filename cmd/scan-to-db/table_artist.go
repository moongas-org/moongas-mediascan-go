package main

import (
	"database/sql"
	"fmt"
	"log"
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

	log.Println("Create artist table...")
	statement, err := db.Prepare(sql) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("artist table created")
}

func insertArtist(db *sql.DB, path string, name string, city string, countrycode string, regioncode string, languagecode string, hasArtistYaml bool) error {
	log.Println("Inserting artist record ...")
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
		log.Fatal(err)
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
			log.Fatal(err)
		}
		log.Println("artist: ", id, " ", path, " ", name, " ", city, " ", countrycode, " ", regioncode, " ", languagecode, " hasartistyaml=", hasArtistYaml)
	}
	if err := row.Err(); err != nil {
		log.Fatal(err)
	}
}
