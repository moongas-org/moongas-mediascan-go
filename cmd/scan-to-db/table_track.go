package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track/streaming"
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
		"albumid" INTEGER REFERENCES album(id),
		"hasspotify" BOOLEAN NOT NULL DEFAULT FALSE,
		"hasmedia" BOOLEAN NOT NULL DEFAULT FALSE,
		"haslyricstxt" BOOLEAN NOT NULL DEFAULT FALSE,
		"haslyricslrc" BOOLEAN NOT NULL DEFAULT FALSE
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

func insertTracks(db *sql.DB, tracks []track.Track, mediaExts []string) error {
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
		hasSpotify := false
		if idData := t.TrackYamlFile.TrackIDData; idData != nil {
			for _, entry := range idData.StreamingData {
				if entry.StreamingService == streaming.Spotify && entry.TrackID != "" {
					hasSpotify = true
					break
				}
			}
		}
		hasMedia := false
		basename := strings.TrimSuffix(t.Path, filepath.Ext(t.Path))
		for _, ext := range mediaExts {
			if regularFileExists(basename + ext) {
				hasMedia = true
				break
			}
		}
		hasLyricsTxt := regularFileExists(basename + ".txt")
		hasLyricsLrc := regularFileExists(basename + ".lrc")
		log.Println("Inserting track record ...")
		_, err = db.Exec(`INSERT INTO track(path, albumpath, artistpath, title, artist, albumartist, album, genre, year, artistid, albumid, hasspotify, hasmedia, haslyricstxt, haslyricslrc)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.Path, albumPath, filepath.Dir(albumPath), data.Title, data.Artist,
			data.AlbumArtist, data.Album, data.Genre, data.Year, artistID, albumID, hasSpotify, hasMedia, hasLyricsTxt, hasLyricsLrc)
		if err != nil {
			return fmt.Errorf("insert track %s: %w", t.Path, err)
		}
	}
	return nil
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("ERROR accessing %s: %v", path, err)
		}
		return false
	}
	return info.Mode().IsRegular()
}

func displayTracks(db *sql.DB) error {
	rows, err := db.Query("SELECT id, path, albumpath, artistpath, title, artist, albumartist, album, genre, year, artistid, albumid, hasspotify, hasmedia, haslyricstxt, haslyricslrc FROM track ORDER BY artist")
	if err != nil {
		return fmt.Errorf("query tracks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, year int
		var artistID, albumID sql.NullInt64
		var path, albumPath, artistPath, title, artist, albumArtist, album, genre string
		var hasSpotify, hasMedia, hasLyricsTxt, hasLyricsLrc bool
		if err := rows.Scan(&id, &path, &albumPath, &artistPath, &title, &artist, &albumArtist, &album, &genre, &year, &artistID, &albumID, &hasSpotify, &hasMedia, &hasLyricsTxt, &hasLyricsLrc); err != nil {
			return fmt.Errorf("scan track row: %w", err)
		}
		log.Println("track: ", id, " ", path, " ", albumPath, " ", artistPath, " ", title, " ", artist, " ", albumArtist, " ", album, " ", genre, " ", year, " artistid=", artistID, " albumid=", albumID, " hasspotify=", hasSpotify, " hasmedia=", hasMedia, " haslyricstxt=", hasLyricsTxt, " haslyricslrc=", hasLyricsLrc)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read tracks: %w", err)
	}
	return nil
}
