package main

// Unlike the yaml scanners, this one scans everything in one cmd,
// i.e. media files, artist.yml files, and album.yml files,
// and outputs a single sqlite database with multiple tables
// (currently the tables are "mediafile", "artist", and "album")

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library

	"github.com/moongas-org/moongas-mediascan-go/internal/shared"
	"gopkg.in/yaml.v3"
)

func main() {
	shared.LogCurrentDir()

	overwriteMode, args, err := shared.ParseCommandArgs(os.Args, 2, 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr, "Usage: go run cmd/scan-to-db <config-yaml> <output-db> [media-root] [--overwrite-existing|--overwrite-newer]")
		fmt.Fprintln(os.Stderr, "Example: go run cmd/scan-to-db mediascan-config.yml mediascan.db /path/to/root --overwrite-existing")
		os.Exit(1)
	}
	configYamlFilepath := args[0]
	outputDBFilepath := args[1]
	mediaRootDir := ""
	if len(args) == 3 {
		mediaRootDir = args[2]
	}

	conf, err := shared.ResolveScanConfig(configYamlFilepath, mediaRootDir)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	outputDBAbsFilepath, err := shared.ResolveOutputAbsPath(outputDBFilepath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(outputDBFilepath); err == nil {
		shouldOverwrite, err := shared.ShouldOverwriteFile(outputDBFilepath, "", overwriteMode)
		if err != nil {
			log.Fatal(err)
		}
		if !shouldOverwrite {
			log.Printf("Skipping existing database %s", outputDBFilepath)
			return
		}
	}

	var files = shared.ScanFiles(conf)
	var artists = shared.ScanArtists(conf)

	os.Remove(outputDBFilepath)

	log.Printf("Creating %s...", outputDBFilepath)
	file, err := os.Create(outputDBFilepath)
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Printf("%s created", outputDBFilepath)

	sqliteDatabase, _ := sql.Open("sqlite3", outputDBFilepath)
	defer sqliteDatabase.Close()
	createTableMediaFile(sqliteDatabase)
	createTableArtist(sqliteDatabase)
	if err := createTableAlbum(sqliteDatabase); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(files); i++ {
		mf := files[i]
		insertMediaFile(sqliteDatabase, mf.Path, mf.AlbumPath, mf.ArtistPath, mf.Title, mf.Artist, mf.AlbumArtist, mf.Album, mf.Genre, mf.Year)
	}

	for i := 0; i < len(artists); i++ {
		a := artists[i]
		insertArtist(sqliteDatabase, a.Path, a.ArtistData.ArtistNames[0], a.ArtistData.City, a.ArtistData.CountryCode, a.ArtistData.RegionCode, a.ArtistData.LanguageCodes[0])
	}

	if err := insertAlbums(sqliteDatabase, files); err != nil {
		log.Fatal(err)
	}

	// DISPLAY INSERTED RECORDS
	displayMediaFiles(sqliteDatabase)
	displayArtists(sqliteDatabase)
	displayAlbums(sqliteDatabase)

	log.Printf("Written to file %s", outputDBAbsFilepath)
}

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

func insertAlbums(db *sql.DB, files []shared.MediaFileData) error {
	seen := make(map[string]bool)
	for _, mf := range files {
		if seen[mf.AlbumPath] {
			continue
		}

		album := shared.AlbumData{
			Name:         mf.Album,
			DateReleased: shared.Date{Y: mf.Year},
		}
		path := filepath.Join(mf.AlbumPath, "album.yml")
		data, err := os.ReadFile(path)
		if err == nil {
			var file shared.AlbumYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			album = file.AlbumData
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read %s: %w", path, err)
		}

		log.Println("Inserting album record ...")
		_, err = db.Exec(`INSERT INTO album(path, name, year, explicitcover) VALUES (?, ?, ?, ?)`,
			mf.AlbumPath, album.Name, album.DateReleased.Y, album.ExplicitCover)
		if err != nil {
			return fmt.Errorf("insert album %s: %w", mf.AlbumPath, err)
		}
		seen[mf.AlbumPath] = true
	}
	return nil
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
