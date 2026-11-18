// Package track scans media libraries for track YAML metadata.
package track

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

// ScanTracks reads root/artist/album/*.yml under each configured media directory.
// Artist and album metadata are excluded; access and decode errors are logged.
func ScanTracks(conf conf.MediascanConf) []track.Track {
	var tracks []track.Track
	for _, mediaDir := range conf.MediaDirs {
		slog.Info("Scanning track media dir", "media_dir", mediaDir)
		err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				slog.Warn("Unable to access path", "path", path, "error", err)
				return nil
			}
			if info == nil {
				slog.Debug("Skipping path with nil file info", "path", path)
				return nil
			}
			relative, err := filepath.Rel(mediaDir, path)
			if err != nil {
				return err
			}
			depth := len(strings.Split(relative, string(filepath.Separator)))
			if info.IsDir() {
				if relative != "." && depth > 2 {
					return filepath.SkipDir
				}
				return nil
			}
			if depth != 3 || filepath.Ext(info.Name()) != ".yml" ||
				info.Name() == "artist.yml" || info.Name() == "album.yml" {
				return nil
			}
			if utils.ContainsAnyOf(path, conf.ExcludePaths) {
				slog.Debug("Skipping excluded path", "path", path)
				return nil
			}
			slog.Debug("Reading filepath", "path", path)
			data, err := os.ReadFile(path)
			if err != nil {
				slog.Warn("Unable to read track metadata", "path", path, "error", err)
				return nil
			}
			var file track.TrackYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				slog.Warn("Unable to parse track metadata", "path", path, "error", err)
				return nil
			}
			tracks = append(tracks, track.Track{TrackYamlFile: file, Path: path})
			return nil
		})
		if err != nil {
			slog.Error("Unable to scan media directory", "media_dir", mediaDir, "error", err)
		}
	}
	return tracks
}
