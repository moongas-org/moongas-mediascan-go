// Package main scans media and writes album metadata as YAML.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"gopkg.in/yaml.v3"
)

func main() {
	cli.LogCurrentDir()

	mode, args, err := cli.ParseCommandArgs(os.Args, 1, 2)
	if err != nil || mode != cli.SkipExisting {
		if err != nil {
			log.Println("Error:", err)
		} else {
			log.Println("Error: album YAML files cannot be overwritten")
		}
		log.Println("Usage: go run ./cmd/scan-to-album-yaml <config-yaml> [media-root]")
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

	if err := writeAlbumYamlFiles(media_files.ScanFiles(conf)); err != nil {
		log.Fatal(err)
	}
}

func writeAlbumYamlFiles(files []media_file.MediaFileData) error {
	genres := make(map[string][]string)
	for i, m := range files {
		if m.Path == nil || *m.Path == "" {
			return fmt.Errorf("media file %d is missing a path", i)
		}
		albumPath := filepath.Dir(*m.Path)
		if m.Genre != nil && *m.Genre != "" && !slices.Contains(genres[albumPath], *m.Genre) {
			genres[albumPath] = append(genres[albumPath], *m.Genre)
		}
	}

	seen := make(map[string]bool)
	countCreated, countSkipped := 0, 0
	for _, m := range files {
		albumPath := filepath.Dir(*m.Path)
		if seen[albumPath] {
			continue
		}
		seen[albumPath] = true

		path := filepath.Join(albumPath, "album.yml")
		albumYaml := album.AlbumYamlFile{
			AlbumData: album.AlbumData{
				Genres:        genres[albumPath],
				DateReleased:  date.Date{},
				ExplicitCover: false,
			},
		}
		if m.Album != nil {
			albumYaml.AlbumData.Name = *m.Album
		}
		if m.Year != nil {
			albumYaml.AlbumData.DateReleased.Y = *m.Year
		}
		data, err := yaml.Marshal(&albumYaml)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", path, err)
		}

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
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
		log.Printf("Created %s", path)
		countCreated++
	}

	log.Printf("Scanned %d media files", len(files))
	log.Printf("Created %d album YAML (.yml) files", countCreated)
	log.Printf("Skipped %d album YAML (.yml) files (already exist)", countSkipped)
	return nil
}
