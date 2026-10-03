package shared

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dhowden/tag"
	"github.com/tcolgate/mp3"
)

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Printf("ERROR: Failed to close file %v\n", err)
		os.Exit(1)
	}
}

func getMp3Duration(path string) (duration float64) {
	t := 0.0
	r, err := os.Open(path)
	if err != nil {
		log.Printf("getMp3Duration error: %v", err)
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
			fmt.Println(err)
			return 0.0
		}
		t = t + f.Duration().Seconds()
	}
	return math.Round(t*100) / 100
}

func ScanFiles(conf MediascanConf) []MediaFileData {
	var files []MediaFileData
	countLoadFailed := 0
	countTagsFailed := 0
	countSkipped := 0
	for _, mediaDir := range conf.MediaDirs {
		if _, err := os.Stat(mediaDir); err != nil {
			log.Printf("ERROR: mediaDir does not exist or could not be accessed: %s (%v)", mediaDir, err)
			continue
		}
		log.Printf("Scanning file media dir: %s", mediaDir)
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

				var m MediaFileData
				m.Path = path                                  // path to media file
				var albumDirPath = filepath.Dir(path)          // path to album dir
				var artistDirPath = filepath.Dir(albumDirPath) // path to artist dir
				m.AlbumPath = albumDirPath
				m.ArtistPath = artistDirPath
				m.Size = info.Size()
				m.ModTime = info.ModTime()
				m.Duration = 0.0
				m.Format = ""
				ext := filepath.Ext(info.Name())
				name := strings.TrimSuffix(info.Name(), ext)
				m.Title = name

				log.Printf("Reading filepath %s", path)
				if info.IsDir() {
					return nil
				}
				if !StringInSlice(ext, conf.MediaExts) {
					return nil
				}

				if ContainsAnyOf(path, conf.ExcludePaths) {
					log.Printf("Skipping %s (ExcludePaths)", path)
					countSkipped += 1
					return nil
				}

				f, err := os.Open(path)
				defer closeFile(f)
				Check(err, path)

				tags, err2 := tag.ReadFrom(f)
				if err2 != nil {
					countTagsFailed += 1
					//log.Printf("ERROR reading tags: %v", err2)
				} else {
					if ContainsAnyOf(tags.Title(), conf.ExcludeTitle) {
						log.Printf("Skipping %s (ExcludeTitle %s)", path, tags.Title())
						countSkipped += 1
						return nil
					}
					if ContainsAnyOf(tags.Artist(), conf.ExcludeArtist) {
						log.Printf("Skipping %s (ExcludeArtist %s)", path, tags.Artist())
						countSkipped += 1
						return nil
					}
					if ContainsAnyOf(tags.AlbumArtist(), conf.ExcludeArtist) {
						log.Printf("Skipping %s (ExcludeAlbumArtist %s)", path, tags.AlbumArtist())
						countSkipped += 1
						return nil
					}
					if ContainsAnyOf(tags.Album(), conf.ExcludeAlbum) {
						log.Printf("Skipping %s (ExcludeAlbum %s)", path, tags.Album())
						countSkipped += 1
						return nil
					}
					if ContainsAnyOf(tags.Genre(), conf.ExcludeGenre) {
						log.Printf("Skipping %s (ExcludeGenre %s)", path, tags.Genre())
						countSkipped += 1
						return nil
					}

					m.Format = string(tags.Format())
					m.Title = tags.Title()
					m.Artist = tags.Artist()
					m.AlbumArtist = tags.AlbumArtist()
					m.Album = tags.Album()
					m.Genre = tags.Genre()
					m.Year = tags.Year()
				}

				if conf.GetMp3Duration && ext == ".mp3" {
					m.Duration = getMp3Duration(path)
				}

				files = append(files, m)

				return nil
			})
		if err != nil {
			log.Printf("ERROR loading file: %v", err)
			countLoadFailed += 1
		}
	}

	switch conf.SortBy {
	case "year":
		sort.SliceStable(files, func(i, j int) bool {
			return files[i].Year < files[j].Year
		})
	case "artist":
		sort.SliceStable(files, func(i, j int) bool {
			return files[i].Artist < files[j].Artist
		})
	}

	log.Printf("Successfully loaded %d media files", len(files))
	if countLoadFailed > 0 {
		log.Printf("Failed to load for %d media files", countLoadFailed)
	}
	if countTagsFailed > 0 {
		log.Printf("Failed to read tags for %d media files", countTagsFailed)
	}
	if countSkipped > 0 {
		log.Printf("Skipped %d media files", countSkipped)
	}
	if countTagsFailed > 0 {
		log.Printf("Failed to load tags for %d media files", countTagsFailed)
	}
	if countSkipped > 0 {
		log.Printf("Skipped %d excluded media files", countSkipped)
	}
	return files
}
