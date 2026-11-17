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

// ScanAlbums returns root/artist/album directories and any other directories
// with readable album.yml files under the configured media directories.
func ScanAlbums(conf conf.MediascanConf) []album.Album {
	var albums []album.Album
	indexByPath := make(map[string]int)
	addAlbum := func(a album.Album) {
		if index, ok := indexByPath[a.Path]; ok {
			if a.HasAlbumYaml {
				albums[index] = a
			}
			return
		}
		indexByPath[a.Path] = len(albums)
		albums = append(albums, a)
	}
	for _, mediaDir := range conf.MediaDirs {
		mediaDir = filepath.Clean(mediaDir)
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
			if info.IsDir() {
				parent := filepath.Dir(path)
				if path != mediaDir && parent != mediaDir && filepath.Dir(parent) == mediaDir &&
					!utils.ContainsAnyOf(path, conf.ExcludePaths) {
					addAlbum(album.Album{Path: path})
				}
				return nil
			}
			if info.Name() != "album.yml" {
				return nil
			}
			if utils.ContainsAnyOf(path, conf.ExcludePaths) {
				log.Printf("Skipping %s (ExcludePaths)", path)
				return nil
			}

			log.Printf("Reading filepath %s", path)
			if info.Mode()&os.ModeSymlink != 0 {
				info, err = os.Stat(path)
				if err != nil {
					log.Printf("ERROR reading %s: %v", path, err)
					return nil
				}
			}
			if !info.Mode().IsRegular() {
				log.Printf("Skipping %s (not a regular file)", path)
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("ERROR reading %s: %v", path, err)
				return nil
			}
			a := album.Album{Path: filepath.Dir(path), HasAlbumYaml: true}
			var file album.AlbumYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				log.Printf("ERROR parsing %s: %v", path, err)
			} else {
				a.AlbumYamlFile = file
			}
			addAlbum(a)
			return nil
		})
		if err != nil {
			log.Printf("ERROR scanning %s: %v", mediaDir, err)
		}
	}
	return albums
}
