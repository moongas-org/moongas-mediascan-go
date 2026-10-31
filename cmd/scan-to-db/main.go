// Package main scans media and stores its metadata in a database.
package main

// Unlike the yaml scanners, this one scans everything in one cmd,
// i.e. media files, artist.yml files, album.yml files, and track YAML files,
// and outputs a single sqlite database with multiple tables
// (currently the tables are "mediafile", "artist", "album", "track", and "trackstreaming")

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	album_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/album"
	artist_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/artist"
	media_file_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	track_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
)

func main() {
	cli.LogCurrentDir()

	overwriteMode, args, err := cli.ParseCommandArgs(os.Args, 2, 3)
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

	conf, err := cli.ResolveScanConfig(configYamlFilepath, mediaRootDir)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	outputDBAbsFilepath, err := cli.ResolveOutputAbsPath(outputDBFilepath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(outputDBFilepath); err == nil {
		shouldOverwrite, err := cli.ShouldOverwriteFile(outputDBFilepath, "", overwriteMode)
		if err != nil {
			log.Fatal(err)
		}
		if !shouldOverwrite {
			log.Printf("Skipping existing database %s", outputDBFilepath)
			return
		}
	}

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
	if err := scanIntoDB(sqliteDatabase, conf); err != nil {
		log.Fatal(err)
	}

	// DISPLAY INSERTED RECORDS
	displayArtists(sqliteDatabase)
	displayAlbums(sqliteDatabase)
	if err := displayTracks(sqliteDatabase); err != nil {
		log.Fatal(err)
	}
	if err := displayTrackStreaming(sqliteDatabase); err != nil {
		log.Fatal(err)
	}
	displayMediaFiles(sqliteDatabase)

	log.Printf("Written to file %s", outputDBAbsFilepath)
}

func scanIntoDB(sqliteDatabase *sql.DB, conf conf.MediascanConf) error {
	createTableArtist(sqliteDatabase)
	if err := createTableAlbum(sqliteDatabase); err != nil {
		return err
	}
	if err := createTableTrack(sqliteDatabase); err != nil {
		return err
	}
	if err := createTableTrackStreaming(sqliteDatabase); err != nil {
		return err
	}
	createTableMediaFile(sqliteDatabase)

	artists := artist_scanner.ScanArtists(conf)
	for i := 0; i < len(artists); i++ {
		a := artists[i]
		data := a.ArtistYamlFile.ArtistData
		insertArtist(sqliteDatabase, a.Path, data.ArtistNames[0], data.City, data.CountryCode, data.RegionCode, data.LanguageCodes[0])
	}

	albums := album_scanner.ScanAlbums(conf)
	if err := insertAlbums(sqliteDatabase, albums); err != nil {
		return err
	}
	tracks := track_scanner.ScanTracks(conf)
	if err := insertTracks(sqliteDatabase, tracks); err != nil {
		return err
	}
	if err := insertTrackStreaming(sqliteDatabase, tracks); err != nil {
		return err
	}
	files := media_file_scanner.ScanFiles(conf)
	return insertMediaFiles(sqliteDatabase, files, tracks)
}
