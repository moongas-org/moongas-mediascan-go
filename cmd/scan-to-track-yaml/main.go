// Package main scans media and writes track metadata as YAML.
package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/logging"
	media_file_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

func main() {
	logging.Init()
	cli.LogCurrentDir()

	overwriteMode, args, err := cli.ParseCommandArgs(os.Args, 1, 2)
	if err != nil {
		slog.Error("Invalid command-line arguments", "error", err)
		slog.Info("Usage: go run cmd/scan-to-track-yaml <config-yaml> [media-root] [--overwrite-existing|--overwrite-newer]")
		slog.Info("Example: go run cmd/scan-to-track-yaml mediascan-config.yml /path/to/root --overwrite-existing")
		os.Exit(1)
	}
	configYamlFilepath := args[0]
	mediaRootDir := ""
	if len(args) == 2 {
		mediaRootDir = args[1]
	}

	conf, err := cli.ResolveScanConfig(configYamlFilepath, mediaRootDir)
	if err != nil {
		slog.Error("Unable to resolve scan configuration", "error", err)
		os.Exit(1)
	}

	files := media_file_scanner.ScanFiles(conf)

	countCreated := 0
	countSkipped := 0
	for _, m := range files {
		ext := filepath.Ext(*m.Path)
		trackYamlFilepath := filepath.Join(filepath.Dir(*m.Path), strings.TrimSuffix(filepath.Base(*m.Path), ext)+".yml")
		if _, err := os.Stat(trackYamlFilepath); err == nil {
			shouldOverwrite, err := cli.ShouldOverwriteFile(trackYamlFilepath, *m.Path, overwriteMode)
			if err != nil {
				slog.Warn("Unable to check whether track YAML should be overwritten", "path", trackYamlFilepath, "error", err)
				continue
			}
			if !shouldOverwrite {
				slog.Debug("Skipping media file because track YAML already exists", "path", *m.Path, "track_yaml", trackYamlFilepath)
				countSkipped += 1
				continue
			}
		}

		trackYaml := track.TrackYamlFile{
			TrackData: track.TrackData{
				Title:       *m.Title,
				Artist:      *m.Artist,
				AlbumArtist: *m.AlbumArtist,
				Album:       *m.Album,
				Genre:       *m.Genre,
				Year:        *m.Year,
			},
		}
		yamlData, err := yaml.Marshal(&trackYaml)
		utils.Check(err, trackYamlFilepath)
		err = os.WriteFile(trackYamlFilepath, yamlData, 0644)
		utils.Check(err, trackYamlFilepath)
		slog.Debug("Created track YAML", "path", trackYamlFilepath)
		countCreated += 1
	}

	slog.Info("Scanned media files", "count", len(files))
	slog.Info("Created track YAML files", "count", countCreated)
	slog.Info("Skipped track YAML files", "count", countSkipped)
}
