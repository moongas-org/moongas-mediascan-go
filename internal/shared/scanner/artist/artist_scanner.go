// Package artist scans media libraries for artist metadata.
package artist

import (
	"log"
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
			log.Printf("ERROR: mediaDir does not exist or could not be accessed: %s (%v)", mediaDir, err)
			continue
		}
		log.Printf("Scanning artist media dir: %s", mediaDir)
		err := filepath.Walk(mediaDir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Printf("ERROR accessing %s: %v", path, err)
					return nil
				}
				if info == nil {
					log.Printf("Skipping nil file info for %s", path)
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
				log.Printf("Reading filepath %s", path)
				fileInfo, err := os.Stat(path)
				if err != nil {
					log.Printf("ERROR inspecting %s: %v", path, err)
					return nil
				}
				if !fileInfo.Mode().IsRegular() {
					log.Printf("Skipping %s (not a regular file)", path)
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					log.Printf("ERROR reading %s: %v", path, err)
					return nil
				}
				a := artist.Artist{Path: filepath.Dir(path), HasArtistYaml: true}
				var metadata artist.ArtistYamlFile
				if err := yaml.Unmarshal(data, &metadata); err != nil {
					log.Printf("ERROR parsing %s: %v", path, err)
				} else {
					a.ArtistYamlFile = metadata
				}
				addArtist(a)
				return nil
			})
		if err != nil {
			log.Printf("ERROR loading file: %v", err)
		}
	}
	return scannedArtists
}
