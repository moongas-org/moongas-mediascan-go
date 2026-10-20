package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared"
	"gopkg.in/yaml.v3"
)

func main() {
	shared.LogCurrentDir()

	mode, args, err := shared.ParseCommandArgs(os.Args, 1, 2)
	if err != nil || mode != shared.SkipExisting {
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

	conf, err := shared.ResolveScanConfig(args[0], mediaRootDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := writeAlbumYamlFiles(shared.ScanFiles(conf)); err != nil {
		log.Fatal(err)
	}
}

func writeAlbumYamlFiles(files []shared.MediaFileData) error {
	seen := make(map[string]bool)
	countCreated, countSkipped := 0, 0
	for _, m := range files {
		if seen[m.AlbumPath] {
			continue
		}
		seen[m.AlbumPath] = true

		path := filepath.Join(m.AlbumPath, "album.yml")
		album := shared.AlbumYamlFile{
			AlbumData: shared.AlbumData{
				Name:          m.Album,
				DateReleased:  shared.Date{Y: m.Year},
				ExplicitCover: false,
			},
		}
		data, err := yaml.Marshal(&album)
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
