package main

import (
	"database/sql"
	"log"
)

func createTableMediaFile(db *sql.DB) {
	sql := `CREATE TABLE mediafile (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"path" TEXT UNIQUE,
		"albumpath" TEXT,
		"artistpath" TEXT,
		"title" TEXT,
		"artist" TEXT,
		"albumartist" TEXT,
		"album" TEXT,
		"genre" TEXT,
		"year" INTEGER
	  );` // SQL Statement for Create Table
	// Keep in sync w/ below: path, albumpath, artistpath, title, artist, albumartist, album, genre, year

	log.Println("Create mediafiles table...")
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// In SQLite, non-unique indexes are created using the standard CREATE INDEX statement,
	// which must be executed after the CREATE TABLE statement.
	sql = `CREATE INDEX idx_artistpath ON mediafile (artistpath);`
	statement, err = db.Prepare(sql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatalln(err.Error())
	}

	sql = `CREATE INDEX idx_albumpath ON mediafile (albumpath);`
	statement, err = db.Prepare(sql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("mediafiles table created")
}

func insertMediaFile(db *sql.DB, path string, albumpath string, artistpath string, title string, artist string, albumartist string, album string, genre string, year int) {
	log.Println("Inserting mediafile record ...")
	sql := `INSERT INTO mediafile(path, albumpath, artistpath, title, artist, albumartist, album, genre, year) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(path, albumpath, artistpath, title, artist, albumartist, album, genre, year)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func displayMediaFiles(db *sql.DB) {
	row, err := db.Query("SELECT * FROM mediafile ORDER BY artist")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		var id int
		var path string
		var albumpath string
		var artistpath string
		var title string
		var artist string
		var albumartist string
		var album string
		var genre string
		var year int
		row.Scan(&id, &path, &albumpath, &artistpath, &title, &artist, &albumartist, &album, &genre, &year)
		log.Println("mediafile: ", id, " ", path, " ", albumpath, " ", artistpath, " ", title, " ", artist, " ", albumartist, " ", album, " ", genre, " ", year)

	}
	if err := row.Err(); err != nil {
		log.Fatal(err)
	}
}
