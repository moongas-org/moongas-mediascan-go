// Package track scans media libraries for track YAML metadata.
package track

import (
	"log"
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
		log.Printf("Scanning track media dir: %s", mediaDir)
		err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("ERROR accessing %s: %v", path, err)
				return nil
			}
			if info == nil {
				log.Printf("Skipping nil file info for %s", path)
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
				log.Printf("Skipping %s (ExcludePaths)", path)
				return nil
			}
			log.Printf("Reading filepath %s", path)
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("ERROR reading %s: %v", path, err)
				return nil
			}
			var file track.TrackYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				log.Printf("ERROR parsing %s: %v", path, err)
				return nil
			}
			tracks = append(tracks, track.Track{TrackYamlFile: file, Path: path})
			return nil
		})
		if err != nil {
			log.Printf("ERROR scanning %s: %v", mediaDir, err)
		}
	}
	return tracks
}
