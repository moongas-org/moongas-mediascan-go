// Package album scans media libraries for album metadata.
package album

import (
	"log/slog"
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
		slog.Info("Scanning album media dir", "media_dir", mediaDir)
		err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				slog.Warn("Unable to access path", "path", path, "error", err)
				return nil
			}
			if info == nil {
				slog.Debug("Skipping path with nil file info", "path", path)
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
				slog.Debug("Skipping excluded path", "path", path)
				return nil
			}

			slog.Debug("Reading filepath", "path", path)
			if info.Mode()&os.ModeSymlink != 0 {
				info, err = os.Stat(path)
				if err != nil {
					slog.Warn("Unable to inspect album metadata", "path", path, "error", err)
					return nil
				}
			}
			if !info.Mode().IsRegular() {
				slog.Debug("Skipping non-regular album metadata file", "path", path)
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				slog.Warn("Unable to read album metadata", "path", path, "error", err)
				return nil
			}
			a := album.Album{Path: filepath.Dir(path), HasAlbumYaml: true}
			var file album.AlbumYamlFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				slog.Warn("Unable to parse album metadata", "path", path, "error", err)
			} else {
				a.AlbumYamlFile = file
			}
			addAlbum(a)
			return nil
		})
		if err != nil {
			slog.Error("Unable to scan media directory", "media_dir", mediaDir, "error", err)
		}
	}
	return albums
}
