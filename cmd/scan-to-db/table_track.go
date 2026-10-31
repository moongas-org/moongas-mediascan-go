package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
)

func createTableTrack(db *sql.DB) error {
	log.Println("Create track table...")
	_, err := db.Exec(`CREATE TABLE track (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"path" TEXT UNIQUE,
		"albumpath" TEXT,
		"artistpath" TEXT,
		"title" TEXT,
		"artist" TEXT,
		"albumartist" TEXT,
		"album" TEXT,
		"genre" TEXT,
		"year" INTEGER,
		"artistid" INTEGER REFERENCES artist(id),
		"albumid" INTEGER REFERENCES album(id)
	  );`)
	if err != nil {
		return fmt.Errorf("create track table: %w", err)
	}
	for _, statement := range []string{
		"CREATE INDEX idx_track_artistpath ON track (artistpath)",
		"CREATE INDEX idx_track_albumpath ON track (albumpath)",
	} {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("create track index: %w", err)
		}
	}
	log.Println("track table created")
	return nil
}

func insertTracks(db *sql.DB, tracks []track.Track) error {
	for _, t := range tracks {
		data := t.TrackYamlFile.TrackData
		albumPath := filepath.Dir(t.Path)
		artistID, err := rowIDByPath(db, "artist", filepath.Dir(albumPath))
		if err != nil {
			return err
		}
		albumID, err := rowIDByPath(db, "album", albumPath)
		if err != nil {
			return err
		}
		log.Println("Inserting track record ...")
		_, err = db.Exec(`INSERT INTO track(path, albumpath, artistpath, title, artist, albumartist, album, genre, year, artistid, albumid)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.Path, albumPath, filepath.Dir(albumPath), data.Title, data.Artist,
			data.AlbumArtist, data.Album, data.Genre, data.Year, artistID, albumID)
		if err != nil {
			return fmt.Errorf("insert track %s: %w", t.Path, err)
		}
	}
	return nil
}

func displayTracks(db *sql.DB) error {
	rows, err := db.Query("SELECT id, path, albumpath, artistpath, title, artist, albumartist, album, genre, year, artistid, albumid FROM track ORDER BY artist")
	if err != nil {
		return fmt.Errorf("query tracks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, year int
		var artistID, albumID sql.NullInt64
		var path, albumPath, artistPath, title, artist, albumArtist, album, genre string
		if err := rows.Scan(&id, &path, &albumPath, &artistPath, &title, &artist, &albumArtist, &album, &genre, &year, &artistID, &albumID); err != nil {
			return fmt.Errorf("scan track row: %w", err)
		}
		log.Println("track: ", id, " ", path, " ", albumPath, " ", artistPath, " ", title, " ", artist, " ", albumArtist, " ", album, " ", genre, " ", year, " artistid=", artistID, " albumid=", albumID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read tracks: %w", err)
	}
	return nil
}
