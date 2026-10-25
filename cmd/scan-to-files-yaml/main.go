// Package main scans media and writes file metadata as YAML.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	media_file_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

func main() {
	cli.LogCurrentDir()

	overwriteMode, args, err := cli.ParseCommandArgs(os.Args, 2, 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr, "Usage: go run cmd/scan-to-files-yaml <config-yaml> <output-yaml> [media-root] [--overwrite-existing|--overwrite-newer]")
		fmt.Fprintln(os.Stderr, "Example: go run cmd/scan-to-files-yaml mediascan-config.yml files.yml /path/to/root --overwrite-existing")
		os.Exit(1)
	}
	configYamlFilepath := args[0]
	outputYamlFilepath := args[1]
	mediaRootDir := ""
	if len(args) == 3 {
		mediaRootDir = args[2]
	}

	conf, err := cli.ResolveScanConfig(configYamlFilepath, mediaRootDir)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	outputYamlAbsFilepath, err := cli.ResolveOutputAbsPath(outputYamlFilepath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(outputYamlFilepath); err == nil {
		shouldOverwrite, err := cli.ShouldOverwriteFile(outputYamlFilepath, "", overwriteMode)
		if err != nil {
			log.Fatal(err)
		}
		if !shouldOverwrite {
			log.Printf("Skipping existing output %s", outputYamlFilepath)
			return
		}
	}

	var files = media_file_scanner.ScanFiles(conf)
	var filesYaml = media_files.MediaFilesYamlFile{}
	filesYaml.Files = files
	yamlData, err := yaml.Marshal(&filesYaml)
	utils.Check(err, "")
	err2 := os.WriteFile(outputYamlFilepath, yamlData, 0644)
	utils.Check(err2, outputYamlFilepath)
	log.Printf("Written to file %s", outputYamlAbsFilepath)
}
