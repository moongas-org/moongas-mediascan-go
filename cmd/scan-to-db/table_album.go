package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
)

func createTableAlbum(db *sql.DB) error {
	log.Println("Create album table...")
	_, err := db.Exec(`CREATE TABLE album (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"path" TEXT UNIQUE,
		"name" TEXT,
		"year" INTEGER,
		"explicitcover" BOOLEAN,
		"artistid" INTEGER REFERENCES artist(id)
	  );`)
	if err != nil {
		return fmt.Errorf("create album table: %w", err)
	}
	log.Println("album table created")
	return nil
}

func insertAlbums(db *sql.DB, albums []album.Album) error {
	seen := make(map[string]bool)
	for i, a := range albums {
		if a.Path == "" {
			return fmt.Errorf("album %d is missing a path", i)
		}
		albumPath := a.Path
		if seen[albumPath] {
			continue
		}

		albumData := a.AlbumYamlFile.AlbumData
		artistID, err := rowIDByPath(db, "artist", filepath.Dir(albumPath))
		if err != nil {
			return err
		}

		log.Println("Inserting album record ...")
		_, err = db.Exec(`INSERT INTO album(path, name, year, explicitcover, artistid) VALUES (?, ?, ?, ?, ?)`,
			albumPath, albumData.Name, albumData.DateReleased.Y, albumData.ExplicitCover, artistID)
		if err != nil {
			return fmt.Errorf("insert album %s: %w", albumPath, err)
		}
		seen[albumPath] = true
	}
	return nil
}

func displayAlbums(db *sql.DB) {
	rows, err := db.Query("SELECT id, path, name, year, explicitcover, artistid FROM album ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var path, name string
		var year int
		var explicitcover bool
		var artistID sql.NullInt64
		if err := rows.Scan(&id, &path, &name, &year, &explicitcover, &artistID); err != nil {
			log.Fatal(err)
		}
		log.Println("album: ", id, " ", path, " ", name, " ", year, " ", explicitcover, " artistid=", artistID)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
