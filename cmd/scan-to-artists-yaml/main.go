package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared"
	"gopkg.in/yaml.v3"
)

func main() {
	shared.LogCurrentDir()

	overwriteMode, args, err := shared.ParseCommandArgs(os.Args, 2, 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr, "Usage: go run cmd/scan-to-artists-yaml <config-yaml> <output-yaml> [media-root] [--overwrite-existing|--overwrite-newer]")
		fmt.Fprintln(os.Stderr, "Example: go run cmd/scan-to-artists-yaml mediascan-config.yml artists.yml /path/to/root --overwrite-existing")
		os.Exit(1)
	}
	configYamlFilepath := args[0]
	outputYamlFilepath := args[1]
	mediaRootDir := ""
	if len(args) == 3 {
		mediaRootDir = args[2]
	}

	conf, err := shared.ResolveScanConfig(configYamlFilepath, mediaRootDir)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	outputYamlAbsFilepath, err := shared.ResolveOutputAbsPath(outputYamlFilepath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(outputYamlFilepath); err == nil {
		shouldOverwrite, err := shared.ShouldOverwriteFile(outputYamlFilepath, "", overwriteMode)
		if err != nil {
			log.Fatal(err)
		}
		if !shouldOverwrite {
			log.Printf("Skipping existing output %s", outputYamlFilepath)
			return
		}
	}

	var artists = shared.ScanArtists(conf)
	var artistsYamlFile = shared.ArtistsYamlFile{}
	artistsYamlFile.Artists = artists

	yamlData, err := yaml.Marshal(&artistsYamlFile)
	shared.Check(err, "")
	err2 := os.WriteFile(outputYamlAbsFilepath, yamlData, 0644)
	shared.Check(err2, outputYamlAbsFilepath)

	log.Printf("Successfully loaded %d artist.yml files", len(artists))
	log.Printf("Written to file %s", outputYamlAbsFilepath)
}
