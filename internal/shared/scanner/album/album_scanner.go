// Package album scans media libraries for album metadata.
package album

import (
	"log"
	"os"
	"path/filepath"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

// ScanAlbums returns album.yml metadata with its containing directory.
// Inaccessible paths and invalid YAML are logged and skipped.
func ScanAlbums(conf conf.MediascanConf) []album.Album {
	var albums []album.Album
	for _, mediaDir := range conf.MediaDirs {
		log.Printf("Scanning album media dir: %s", mediaDir)
		err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("ERROR accessing %s: %v", path, err)
				return nil
			}
			if info == nil {
				log.Printf("Skipping nil file info for %s", path)
				return nil
			}
			if info.IsDir() || info.Name() != "album.yml" {
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
			var file album.AlbumYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				log.Printf("ERROR parsing %s: %v", path, err)
				return nil
			}
			albums = append(albums, album.Album{
				AlbumYamlFile: file,
				Path:          filepath.Dir(path),
			})
			return nil
		})
		if err != nil {
			log.Printf("ERROR scanning %s: %v", mediaDir, err)
		}
	}
	return albums
}
