package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"gopkg.in/yaml.v3"
)

func createTableAlbum(db *sql.DB) error {
	log.Println("Create album table...")
	_, err := db.Exec(`CREATE TABLE album (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"path" TEXT UNIQUE,
		"name" TEXT,
		"year" INTEGER,
		"explicitcover" BOOLEAN
	  );`)
	if err != nil {
		return fmt.Errorf("create album table: %w", err)
	}
	log.Println("album table created")
	return nil
}

func insertAlbums(db *sql.DB, files []media_file.MediaFileData) error {
	seen := make(map[string]bool)
	for _, mf := range files {
		if seen[mf.AlbumPath] {
			continue
		}

		albumData := album.AlbumData{
			DateReleased: date.Date{},
		}
		if mf.Album != nil {
			albumData.Name = *mf.Album
		}
		if mf.Year != nil {
			albumData.DateReleased.Y = *mf.Year
		}
		path := filepath.Join(mf.AlbumPath, "album.yml")
		data, err := os.ReadFile(path)
		if err == nil {
			var file album.AlbumYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			albumData = file.AlbumData
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read %s: %w", path, err)
		}

		log.Println("Inserting album record ...")
		_, err = db.Exec(`INSERT INTO album(path, name, year, explicitcover) VALUES (?, ?, ?, ?)`,
			mf.AlbumPath, albumData.Name, albumData.DateReleased.Y, albumData.ExplicitCover)
		if err != nil {
			return fmt.Errorf("insert album %s: %w", mf.AlbumPath, err)
		}
		seen[mf.AlbumPath] = true
	}
	return nil
}

func displayAlbums(db *sql.DB) {
	rows, err := db.Query("SELECT id, path, name, year, explicitcover FROM album ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var path, name string
		var year int
		var explicitcover bool
		if err := rows.Scan(&id, &path, &name, &year, &explicitcover); err != nil {
			log.Fatal(err)
		}
		log.Println("album: ", id, " ", path, " ", name, " ", year, " ", explicitcover)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
