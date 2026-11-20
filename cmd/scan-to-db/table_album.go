package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
)

func createTableAlbum(db *sql.DB) error {
	slog.Info("Creating album table")
	_, err := db.Exec(`CREATE TABLE album (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"path" TEXT UNIQUE,
		"name" TEXT,
		"year" INTEGER,
		"explicitcover" BOOLEAN,
		"artistid" INTEGER REFERENCES artist(id),
		"genre" TEXT,
		"hasalbumyaml" BOOLEAN NOT NULL DEFAULT FALSE,
		"hascover" BOOLEAN NOT NULL DEFAULT FALSE
	  );`)
	if err != nil {
		return fmt.Errorf("create album table: %w", err)
	}
	slog.Info("Created album table")
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
		name := albumData.Name
		if name == "" {
			name = filepath.Base(albumPath)
		}
		genre := ""
		if len(albumData.Genres) > 0 {
			genre = albumData.Genres[0]
		}
		artistID, err := rowIDByPath(db, "artist", filepath.Dir(albumPath))
		if err != nil {
			return err
		}

		hasCover := coverImageExists(filepath.Join(albumPath, "cover.jpg")) ||
			coverImageExists(filepath.Join(albumPath, "cover.webp"))

		slog.Debug("Inserting album record", "path", albumPath)
		_, err = db.Exec(`INSERT INTO album(path, name, year, explicitcover, artistid, genre, hasalbumyaml, hascover) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			albumPath, name, albumData.DateReleased.Y, albumData.ExplicitCover, artistID, genre, a.HasAlbumYaml, hasCover)
		if err != nil {
			return fmt.Errorf("insert album %s: %w", albumPath, err)
		}
		seen[albumPath] = true
	}
	return nil
}

func coverImageExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("Unable to access cover image file", "path", path, "error", err)
		}
		return false
	}
	return info.Mode().IsRegular()
}

func displayAlbums(db *sql.DB) error {
	rows, err := db.Query("SELECT id, path, name, year, explicitcover, artistid, genre, hasalbumyaml, hascover FROM album ORDER BY name")
	if err != nil {
		return fmt.Errorf("query albums: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var path, name, genre string
		var year int
		var explicitcover, hasAlbumYaml, hasCover bool
		var artistID sql.NullInt64
		if err := rows.Scan(&id, &path, &name, &year, &explicitcover, &artistID, &genre, &hasAlbumYaml, &hasCover); err != nil {
			return fmt.Errorf("scan album row: %w", err)
		}
		slog.Debug("Album row", "id", id, "path", path, "name", name, "year", year, "explicit_cover", explicitcover, "artist_id", artistID, "genre", genre, "has_album_yaml", hasAlbumYaml, "has_cover", hasCover)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read album rows: %w", err)
	}
	return nil
}
