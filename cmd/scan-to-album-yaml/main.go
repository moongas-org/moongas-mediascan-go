// Package main scans media and writes album metadata as YAML.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	track_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"gopkg.in/yaml.v3"
)

func main() {
	cli.LogCurrentDir()

	mode, args, err := cli.ParseCommandArgs(os.Args, 1, 2)
	if err != nil || mode == cli.OverwriteNewer {
		if err != nil {
			log.Println("Error:", err)
		} else {
			log.Println("Error: --overwrite-newer is not supported; use --overwrite-existing")
		}
		log.Println("Usage: go run ./cmd/scan-to-album-yaml <config-yaml> [media-root] [--overwrite-existing]")
		log.Println("Existing album.yml files are preserved by default. Add --overwrite-existing to replace them with generated metadata.")
		os.Exit(1)
	}
	mediaRootDir := ""
	if len(args) == 2 {
		mediaRootDir = args[1]
	}

	conf, err := cli.ResolveScanConfig(args[0], mediaRootDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := scanToAlbumYaml(conf, mode); err != nil {
		log.Fatal(err)
	}
}

func scanToAlbumYaml(conf conf.MediascanConf, mode cli.OverwriteMode) error {
	return writeAlbumYamlFiles(media_files.ScanFiles(conf), track_scanner.ScanTracks(conf), mode)
}

func writeAlbumYamlFiles(files []media_file.MediaFileData, tracks []track.Track, mode cli.OverwriteMode) error {
	if mode != cli.SkipExisting && mode != cli.OverwriteExisting {
		return fmt.Errorf("unsupported album YAML overwrite mode %d", mode)
	}

	albums := make(map[string]*album.AlbumData)
	var albumPaths []string
	addMetadata := func(dir, name string, year int, genre string) {
		data := albums[dir]
		if data == nil {
			data = &album.AlbumData{Name: name, DateReleased: date.Date{Y: year}}
			albums[dir] = data
			albumPaths = append(albumPaths, dir)
		}
		if genre != "" && !slices.Contains(data.Genres, genre) {
			data.Genres = append(data.Genres, genre)
		}
	}

	// Process YAML first so it supplies album-level fields and replaces tags
	// from the media files it describes, without dropping unrelated media genres.
	mediaPaths, sidecarStems := make(map[string]bool), make(map[string]bool)
	for i, t := range tracks {
		if t.Path == "" {
			return fmt.Errorf("track YAML file %d is missing a path", i)
		}
		dir := filepath.Dir(t.Path)
		data := t.TrackYamlFile.TrackData
		addMetadata(dir, data.Album, data.Year, data.Genre)

		hasMediaPath := false
		for _, m := range t.TrackYamlFile.MediaFileData {
			if m.Path == nil || *m.Path == "" {
				continue
			}
			hasMediaPath = true
			path := filepath.Clean(*m.Path)
			if !filepath.IsAbs(path) {
				path = filepath.Join(dir, path)
			}
			if filepath.Dir(path) == dir {
				mediaPaths[path] = true
			}
		}
		if !hasMediaPath {
			sidecarStems[fileStem(filepath.Clean(t.Path))] = true
		}
	}
	for i, m := range files {
		if m.Path == nil || *m.Path == "" {
			return fmt.Errorf("media file %d is missing a path", i)
		}
		path := filepath.Clean(*m.Path)
		if mediaPaths[path] || sidecarStems[fileStem(path)] {
			continue
		}
		addMetadata(filepath.Dir(path), valueOrZero(m.Album), valueOrZero(m.Year), valueOrZero(m.Genre))
	}

	countCreated, countSkipped := 0, 0
	for _, albumPath := range albumPaths {
		path := filepath.Join(albumPath, "album.yml")
		explicitCover := false
		if mode == cli.OverwriteExisting {
			existingData, err := os.ReadFile(path)
			if err == nil {
				var existing album.AlbumYamlFile
				if err := yaml.Unmarshal(existingData, &existing); err != nil {
					return fmt.Errorf("parse existing %s to preserve explicitCover: %w", path, err)
				}
				explicitCover = existing.AlbumData.ExplicitCover
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("read existing %s to preserve explicitCover: %w", path, err)
			}
		}
		albumData := *albums[albumPath]
		albumData.ExplicitCover = explicitCover
		albumYaml := album.AlbumYamlFile{AlbumData: albumData}
		data, err := yaml.Marshal(&albumYaml)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", path, err)
		}

		flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if mode == cli.OverwriteExisting {
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}
		file, err := os.OpenFile(path, flags, 0o644)
		if os.IsExist(err) {
			log.Printf("Skipping %s (already exists)", path)
			countSkipped++
			continue
		}
		if err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		_, writeErr := file.Write(data)
		closeErr := file.Close()
		if writeErr != nil {
			return fmt.Errorf("write %s: %w", path, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", path, closeErr)
		}
		log.Printf("Written %s", path)
		countCreated++
	}

	log.Printf("Scanned %d media files", len(files))
	log.Printf("Scanned %d track YAML (.yml) files", len(tracks))
	log.Printf("Written %d album YAML (.yml) files", countCreated)
	log.Printf("Skipped %d album YAML (.yml) files (already exist)", countSkipped)
	return nil
}

func fileStem(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(filepath.Base(path)))
}

func valueOrZero[T any](value *T) T {
	if value != nil {
		return *value
	}
	var zero T
	return zero
}
