package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
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
		"year" INTEGER,
		"size" INTEGER,
		"format" TEXT,
		"duration" REAL,
		"artistid" INTEGER REFERENCES artist(id),
		"albumid" INTEGER REFERENCES album(id),
		"trackid" INTEGER REFERENCES track(id)
	  );` // SQL Statement for Create Table

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

func insertMediaFile(db *sql.DB, path string, albumpath string, artistpath string, title string, artist string, albumartist string, album string, genre string, year int, size *int64, format *string, duration *float64, artistID, albumID, trackID sql.NullInt64) error {
	log.Println("Inserting mediafile record ...")
	_, err := db.Exec(`INSERT INTO mediafile(path, albumpath, artistpath, title, artist, albumartist, album, genre, year, size, format, duration, artistid, albumid, trackid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		path, albumpath, artistpath, title, artist, albumartist, album, genre, year, size, format, duration, artistID, albumID, trackID)
	if err != nil {
		return fmt.Errorf("insert mediafile %s: %w", path, err)
	}
	return nil
}

func insertMediaFiles(db *sql.DB, files []media_file.MediaFileData, tracks []track.Track) error {
	trackPaths := make(map[string]string)
	for _, t := range tracks {
		for _, mf := range t.TrackYamlFile.MediaFileData {
			if mf.Path == nil || *mf.Path == "" {
				log.Printf("WARNING: mediaFileData entry in %s has no path; cannot link a media file", t.Path)
				continue
			}
			path := *mf.Path
			if !filepath.IsAbs(path) {
				path = filepath.Join(filepath.Dir(t.Path), path)
			}
			path, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve media-file path in %s: %w", t.Path, err)
			}
			if previous, ok := trackPaths[path]; ok && previous != t.Path {
				return fmt.Errorf("media file %s is claimed by both %s and %s", path, previous, t.Path)
			}
			trackPaths[path] = t.Path
		}
	}
	for i, mf := range files {
		if mf.Path == nil || *mf.Path == "" {
			return fmt.Errorf("media file %d is missing a path", i)
		}
		path, err := filepath.Abs(*mf.Path)
		if err != nil {
			return fmt.Errorf("resolve media-file path %s: %w", *mf.Path, err)
		}
		albumPath := filepath.Dir(*mf.Path)
		artistPath := filepath.Dir(albumPath)
		artistID, err := rowIDByPath(db, "artist", artistPath)
		if err != nil {
			return err
		}
		albumID, err := rowIDByPath(db, "album", albumPath)
		if err != nil {
			return err
		}
		var trackID sql.NullInt64
		if trackPath, ok := trackPaths[path]; ok {
			trackID, err = rowIDByPath(db, "track", trackPath)
			if err != nil {
				return err
			}
		} else {
			log.Printf("WARNING: no track mediaFileData path matches %s; storing NULL trackid", *mf.Path)
		}
		if err := insertMediaFile(db, *mf.Path, albumPath, artistPath, *mf.Title, *mf.Artist,
			*mf.AlbumArtist, *mf.Album, *mf.Genre, *mf.Year, mf.Size, mf.Format, mf.Duration, artistID, albumID, trackID); err != nil {
			return err
		}
	}
	return nil
}

func displayMediaFiles(db *sql.DB) {
	row, err := db.Query("SELECT id, path, albumpath, artistpath, title, artist, albumartist, album, genre, year, size, format, duration, artistid, albumid, trackid FROM mediafile ORDER BY artist")
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
		var size sql.NullInt64
		var format sql.NullString
		var duration sql.NullFloat64
		var artistID, albumID, trackID sql.NullInt64
		if err := row.Scan(&id, &path, &albumpath, &artistpath, &title, &artist, &albumartist, &album, &genre, &year,
			&size, &format, &duration, &artistID, &albumID, &trackID); err != nil {
			log.Fatal(err)
		}
		log.Println("mediafile: ", id, " ", path, " ", albumpath, " ", artistpath, " ", title, " ", artist, " ", albumartist, " ", album, " ", genre, " ", year,
			" size=", size, " format=", format, " duration=", duration, " artistid=", artistID, " albumid=", albumID, " trackid=", trackID)

	}
	if err := row.Err(); err != nil {
		log.Fatal(err)
	}
}
