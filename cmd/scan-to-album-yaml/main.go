// Package main scans media and writes album metadata as YAML.
package main

import (
	"errors"
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

	if err := writeAlbumYamlFiles(media_files.ScanFiles(conf), mode); err != nil {
		log.Fatal(err)
	}
}

func writeAlbumYamlFiles(files []media_file.MediaFileData, mode cli.OverwriteMode) error {
	if mode != cli.SkipExisting && mode != cli.OverwriteExisting {
		return fmt.Errorf("unsupported album YAML overwrite mode %d", mode)
	}
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
		albumYaml := album.AlbumYamlFile{
			AlbumData: album.AlbumData{
				Genres:        genres[albumPath],
				DateReleased:  date.Date{},
				ExplicitCover: explicitCover,
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
	log.Printf("Written %d album YAML (.yml) files", countCreated)
	log.Printf("Skipped %d album YAML (.yml) files (already exist)", countSkipped)
	return nil
}
