package main

import (
	"database/sql"
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
		"languagecode" TEXT
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

func insertArtist(db *sql.DB, path string, name string, city string, countrycode string, regioncode string, languagecode string) {
	log.Println("Inserting artist record ...")
	insertStudentSQL := `INSERT INTO artist(path, name, city, countrycode, regioncode, languagecode) VALUES (?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertStudentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(path, name, city, countrycode, regioncode, languagecode)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func displayArtists(db *sql.DB) {
	row, err := db.Query("SELECT * FROM artist ORDER BY name")
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
		row.Scan(&id, &path, &name, &city, &countrycode, &regioncode, &languagecode)
		log.Println("artist: ", id, " ", path, " ", name, " ", city, " ", countrycode, " ", regioncode, " ", languagecode)
	}
	if err := row.Err(); err != nil {
		log.Fatal(err)
	}
}
