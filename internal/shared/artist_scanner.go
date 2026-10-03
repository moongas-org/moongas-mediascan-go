package shared

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ScanArtists(conf MediascanConf) []ArtistsYamlArtistData {
	var artists []ArtistsYamlArtistData
	countLoadFailed := 0
	countSkipped := 0
	for _, mediaDir := range conf.MediaDirs {
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

				if !info.IsDir() && info.Name() == "artist.yml" {
					log.Printf("Reading filepath %s", path)
					if ContainsAnyOf(path, conf.ExcludePaths) {
						log.Printf("Skipping %s (ExcludePaths)", path)
						countSkipped += 1
						return nil
					}
					var a ArtistsYamlArtistData
					a.Path = filepath.Dir(path)

					// Read the YAML file
					data, err := os.ReadFile(path)
					Check(err, path)

					// Unmarshal the YAML data into the Config struct
					var dataFile ArtistYamlFile
					err = yaml.Unmarshal(data, &dataFile)
					Check(err, path)

					a.ArtistData = dataFile.ArtistData

					artists = append(artists, a)
				}
				return nil
			})
		if err != nil {
			log.Printf("ERROR loading file: %v", err)
			countLoadFailed += 1
		}
	}
	return artists
}
