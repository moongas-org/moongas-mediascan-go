// Package main scans media and writes track metadata as YAML.
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	media_file_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

func main() {
	cli.LogCurrentDir()

	overwriteMode, args, err := cli.ParseCommandArgs(os.Args, 1, 2)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Usage: go run cmd/scan-to-track-yaml <config-yaml> [media-root] [--overwrite-existing|--overwrite-newer]")
		log.Println("Example: go run cmd/scan-to-track-yaml mediascan-config.yml /path/to/root --overwrite-existing")
		os.Exit(1)
	}
	configYamlFilepath := args[0]
	mediaRootDir := ""
	if len(args) == 2 {
		mediaRootDir = args[1]
	}

	conf, err := cli.ResolveScanConfig(configYamlFilepath, mediaRootDir)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	files := media_file_scanner.ScanFiles(conf)

	countCreated := 0
	countSkipped := 0
	for _, m := range files {
		ext := filepath.Ext(m.Path)
		trackYamlFilepath := filepath.Join(m.AlbumPath, strings.TrimSuffix(filepath.Base(m.Path), ext)+".yml")
		if _, err := os.Stat(trackYamlFilepath); err == nil {
			shouldOverwrite, err := cli.ShouldOverwriteFile(trackYamlFilepath, m.Path, overwriteMode)
			if err != nil {
				log.Printf("ERROR: %v", err)
				continue
			}
			if !shouldOverwrite {
				log.Printf("Skipping %s (%s already exists)", m.Path, trackYamlFilepath)
				countSkipped += 1
				continue
			}
		}

		trackYaml := track.TrackYamlFile{
			TrackData: track.TrackData{
				Title:       m.Title,
				Artist:      m.Artist,
				AlbumArtist: m.AlbumArtist,
				Album:       m.Album,
				Genre:       m.Genre,
				Year:        m.Year,
			},
		}
		yamlData, err := yaml.Marshal(&trackYaml)
		utils.Check(err, trackYamlFilepath)
		err = os.WriteFile(trackYamlFilepath, yamlData, 0644)
		utils.Check(err, trackYamlFilepath)
		log.Printf("Created %s", trackYamlFilepath)
		countCreated += 1
	}

	log.Printf("Scanned %d media files", len(files))
	log.Printf("Created %d track YAML (.yml) files", countCreated)
	log.Printf("Skipped %d track YAML (.yml) files (already exist)", countSkipped)
}
