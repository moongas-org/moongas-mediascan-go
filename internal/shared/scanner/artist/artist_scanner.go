// Package artist scans media libraries for artist metadata.
package artist

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/artist"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

// ScanArtists returns direct child artist directories and any directories with
// readable artist.yml files elsewhere under the configured media directories.
func ScanArtists(conf conf.MediascanConf) []artist.Artist {
	var scannedArtists []artist.Artist
	indexByPath := make(map[string]int)
	addArtist := func(a artist.Artist) {
		if index, ok := indexByPath[a.Path]; ok {
			if a.HasArtistYaml {
				scannedArtists[index] = a
			}
			return
		}
		indexByPath[a.Path] = len(scannedArtists)
		scannedArtists = append(scannedArtists, a)
	}
	for _, mediaDir := range conf.MediaDirs {
		mediaDir = filepath.Clean(mediaDir)
		if _, err := os.Stat(mediaDir); err != nil {
			slog.Warn("Media directory does not exist or could not be accessed", "media_dir", mediaDir, "error", err)
			continue
		}
		slog.Info("Scanning artist media dir", "media_dir", mediaDir)
		err := filepath.Walk(mediaDir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					slog.Warn("Unable to access path", "path", path, "error", err)
					return nil
				}
				if info == nil {
					slog.Debug("Skipping path with nil file info", "path", path)
					return nil
				}

				if info.IsDir() {
					if path != mediaDir && filepath.Dir(path) == mediaDir &&
						!utils.ContainsAnyOf(path, conf.ExcludePaths) {
						addArtist(artist.Artist{Path: path})
					}
					return nil
				}
				if info.Name() != "artist.yml" || utils.ContainsAnyOf(path, conf.ExcludePaths) {
					return nil
				}
				slog.Debug("Reading filepath", "path", path)
				fileInfo, err := os.Stat(path)
				if err != nil {
					slog.Warn("Unable to inspect artist metadata", "path", path, "error", err)
					return nil
				}
				if !fileInfo.Mode().IsRegular() {
					slog.Debug("Skipping non-regular artist metadata file", "path", path)
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					slog.Warn("Unable to read artist metadata", "path", path, "error", err)
					return nil
				}
				a := artist.Artist{Path: filepath.Dir(path), HasArtistYaml: true}
				var metadata artist.ArtistYamlFile
				if err := yaml.Unmarshal(data, &metadata); err != nil {
					slog.Warn("Unable to parse artist metadata", "path", path, "error", err)
				} else {
					a.ArtistYamlFile = metadata
				}
				addArtist(a)
				return nil
			})
		if err != nil {
			slog.Error("Unable to scan artist metadata", "error", err)
		}
	}
	return scannedArtists
}
