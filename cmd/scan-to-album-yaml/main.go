// Package main scans media and writes album metadata as YAML.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
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
	seen := make(map[string]bool)
	countCreated, countSkipped := 0, 0
	for _, m := range files {
		if seen[m.AlbumPath] {
			continue
		}
		seen[m.AlbumPath] = true

		path := filepath.Join(m.AlbumPath, "album.yml")
		albumYaml := album.AlbumYamlFile{
			AlbumData: album.AlbumData{
				Name:          m.Album,
				DateReleased:  types.Date{Y: m.Year},
				ExplicitCover: false,
			},
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
