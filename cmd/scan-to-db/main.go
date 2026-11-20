// Package main scans media and stores its metadata in a database.
package main

// Unlike the yaml scanners, this one scans everything in one cmd,
// i.e. media files, artist.yml files, album.yml files, and track YAML files,
// and outputs a single sqlite database with multiple tables
// (currently the tables are "mediafile", "artist", "album", "track", and "trackstreaming")

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/logging"
	album_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/album"
	artist_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/artist"
	media_file_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	track_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
)

func main() {
	logging.Init()
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
		slog.Error("Unable to resolve scan configuration", "error", err)
		os.Exit(1)
	}

	outputDBAbsFilepath, err := cli.ResolveOutputAbsPath(outputDBFilepath)
	if err != nil {
		slog.Error("Unable to resolve output path", "error", err)
		os.Exit(1)
	}

	if err := writeDatabase(outputDBAbsFilepath, overwriteMode, conf); err != nil {
		slog.Error("Unable to build database", "path", outputDBAbsFilepath, "error", err)
		os.Exit(1)
	}
}

func writeDatabase(outputPath string, overwriteMode cli.OverwriteMode, conf conf.MediascanConf) error {
	shouldOverwrite, err := cli.ShouldOverwriteFile(outputPath, "", overwriteMode)
	if err != nil {
		return fmt.Errorf("check whether database should be overwritten: %w", err)
	}
	if !shouldOverwrite {
		slog.Info("Skipping existing database", "path", outputPath)
		return nil
	}

	stagingDir, err := os.MkdirTemp(filepath.Dir(outputPath), ".scan-to-db-*")
	if err != nil {
		return fmt.Errorf("create database staging directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(stagingDir); err != nil {
			slog.Warn("Unable to clean up database staging directory", "path", stagingDir, "error", err)
		}
	}()
	stagedPath := filepath.Join(stagingDir, "scan.db")
	slog.Info("Building database before publishing", "path", outputPath)
	if err := buildDatabase(stagedPath, conf); err != nil {
		return err
	}

	if overwriteMode == cli.SkipExisting {
		// Do not clobber a database published by another scan while this one was running.
		err = os.Link(stagedPath, outputPath)
	} else {
		err = os.Rename(stagedPath, outputPath)
	}
	if err != nil {
		return fmt.Errorf("publish database: %w", err)
	}
	slog.Info("Wrote database file", "path", outputPath)
	return nil
}

func buildDatabase(path string, conf conf.MediascanConf) (err error) {
	dsn := url.URL{Scheme: "file", Path: path}
	db, err := sql.Open("sqlite3", dsn.String())
	if err != nil {
		return fmt.Errorf("open staged database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close staged database: %w", closeErr)
			} else {
				slog.Warn("Unable to close staged database", "path", path, "error", closeErr)
			}
		}
	}()

	if err := scanIntoDB(db, conf); err != nil {
		return fmt.Errorf("scan media into database: %w", err)
	}
	for _, display := range []func(*sql.DB) error{
		displayArtists, displayAlbums, displayTracks, displayTrackStreaming, displayMediaFiles,
	} {
		if err := display(db); err != nil {
			return err
		}
	}
	return nil
}

func scanIntoDB(sqliteDatabase *sql.DB, conf conf.MediascanConf) error {
	for _, createTable := range []func(*sql.DB) error{
		createTableArtist, createTableAlbum, createTableTrack, createTableTrackStreaming, createTableMediaFile,
	} {
		if err := createTable(sqliteDatabase); err != nil {
			return err
		}
	}

	artists := artist_scanner.ScanArtists(conf)
	for _, a := range artists {
		data := a.ArtistYamlFile.ArtistData
		name := filepath.Base(a.Path)
		if len(data.ArtistNames) > 0 && data.ArtistNames[0] != "" {
			name = data.ArtistNames[0]
		}
		var language string
		if len(data.LanguageCodes) > 0 {
			language = data.LanguageCodes[0]
		}
		if err := insertArtist(sqliteDatabase, a.Path, name, data.City, data.CountryCode, data.RegionCode, language, a.HasArtistYaml); err != nil {
			return err
		}
	}

	albums := album_scanner.ScanAlbums(conf)
	if err := insertAlbums(sqliteDatabase, albums); err != nil {
		return err
	}
	tracks := track_scanner.ScanTracks(conf)
	if err := insertTracks(sqliteDatabase, tracks, conf.MediaExts); err != nil {
		return err
	}
	if err := insertTrackStreaming(sqliteDatabase, tracks); err != nil {
		return err
	}
	files := media_file_scanner.ScanFiles(conf)
	return insertMediaFiles(sqliteDatabase, files, tracks)
}
