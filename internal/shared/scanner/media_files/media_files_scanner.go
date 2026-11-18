// Package media_files scans media libraries for audio files.
package media_files

import (
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dhowden/tag"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"github.com/tcolgate/mp3"
)

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		slog.Error("Failed to close file", "error", err)
		os.Exit(1)
	}
}

func getMp3Duration(path string) (duration float64) {
	t := 0.0
	r, err := os.Open(path)
	if err != nil {
		slog.Warn("Unable to open MP3 to determine duration", "path", path, "error", err)
		return 0.0
	}
	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0
	for {
		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			slog.Warn("Unable to decode MP3 frame", "path", path, "error", err)
			return 0.0
		}
		t = t + f.Duration().Seconds()
	}
	return math.Round(t*100) / 100
}

func ScanFiles(conf conf.MediascanConf) []media_file.MediaFileData {
	var files []media_file.MediaFileData
	countLoadFailed := 0
	countTagsFailed := 0
	countSkipped := 0
	for _, mediaDir := range conf.MediaDirs {
		if _, err := os.Stat(mediaDir); err != nil {
			slog.Warn("Media directory does not exist or could not be accessed", "media_dir", mediaDir, "error", err)
			continue
		}
		slog.Info("Scanning file media dir", "media_dir", mediaDir)
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

				var m media_file.MediaFileData
				m.Path = utils.Pointer(path)
				m.Size = utils.Pointer(info.Size())
				m.ModTime = utils.Pointer(info.ModTime())
				m.Duration = utils.Pointer(0.0)
				m.Format = utils.Pointer("")
				m.Artist = utils.Pointer("")
				m.AlbumArtist = utils.Pointer("")
				m.Album = utils.Pointer("")
				m.Genre = utils.Pointer("")
				m.Year = utils.Pointer(0)
				ext := filepath.Ext(info.Name())
				name := strings.TrimSuffix(info.Name(), ext)
				m.Title = utils.Pointer(name)

				slog.Debug("Reading filepath", "path", path)
				if info.IsDir() {
					return nil
				}
				if !utils.StringInSlice(ext, conf.MediaExts) {
					return nil
				}

				if utils.ContainsAnyOf(path, conf.ExcludePaths) {
					slog.Debug("Skipping excluded path", "path", path)
					countSkipped += 1
					return nil
				}

				f, err := os.Open(path)
				defer closeFile(f)
				utils.Check(err, path)

				tags, err2 := tag.ReadFrom(f)
				if err2 != nil {
					countTagsFailed += 1
					slog.Debug("Unable to read media tags", "path", path, "error", err2)
				} else {
					if utils.ContainsAnyOf(tags.Title(), conf.ExcludeTitle) {
						slog.Debug("Skipping media file excluded by title", "path", path, "title", tags.Title())
						countSkipped += 1
						return nil
					}
					if utils.ContainsAnyOf(tags.Artist(), conf.ExcludeArtist) {
						slog.Debug("Skipping media file excluded by artist", "path", path, "artist", tags.Artist())
						countSkipped += 1
						return nil
					}
					if utils.ContainsAnyOf(tags.AlbumArtist(), conf.ExcludeArtist) {
						slog.Debug("Skipping media file excluded by album artist", "path", path, "album_artist", tags.AlbumArtist())
						countSkipped += 1
						return nil
					}
					if utils.ContainsAnyOf(tags.Album(), conf.ExcludeAlbum) {
						slog.Debug("Skipping media file excluded by album", "path", path, "album", tags.Album())
						countSkipped += 1
						return nil
					}
					if utils.ContainsAnyOf(tags.Genre(), conf.ExcludeGenre) {
						slog.Debug("Skipping media file excluded by genre", "path", path, "genre", tags.Genre())
						countSkipped += 1
						return nil
					}

					m.Format = utils.Pointer(string(tags.Format()))
					m.Title = utils.Pointer(tags.Title())
					m.Artist = utils.Pointer(tags.Artist())
					m.AlbumArtist = utils.Pointer(tags.AlbumArtist())
					m.Album = utils.Pointer(tags.Album())
					m.Genre = utils.Pointer(tags.Genre())
					m.Year = utils.Pointer(tags.Year())
				}

				if conf.GetMp3Duration && ext == ".mp3" {
					m.Duration = utils.Pointer(getMp3Duration(path))
				}

				files = append(files, m)

				return nil
			})
		if err != nil {
			slog.Error("Unable to scan media directory", "error", err)
			countLoadFailed += 1
		}
	}

	switch conf.SortBy {
	case "year":
		sort.SliceStable(files, func(i, j int) bool {
			return *files[i].Year < *files[j].Year
		})
	case "artist":
		sort.SliceStable(files, func(i, j int) bool {
			return *files[i].Artist < *files[j].Artist
		})
	}

	slog.Info("Successfully loaded media files", "count", len(files))
	if countLoadFailed > 0 {
		slog.Warn("Failed to load media files", "count", countLoadFailed)
	}
	if countTagsFailed > 0 {
		slog.Warn("Failed to read tags for media files", "count", countTagsFailed)
	}
	if countSkipped > 0 {
		slog.Info("Skipped media files", "count", countSkipped)
	}
	if countTagsFailed > 0 {
		slog.Warn("Failed to load tags for media files", "count", countTagsFailed)
	}
	if countSkipped > 0 {
		slog.Info("Skipped excluded media files", "count", countSkipped)
	}
	return files
}
