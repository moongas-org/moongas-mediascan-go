package shared

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadConf(configYamlFilepath string) (conf MediascanConf) {
	configYamlAbsFilepath, err := filepath.Abs(configYamlFilepath)
	if err != nil {
		log.Fatalf("Failed to get absolute path to yaml config file: %v", err)
	} else {
		log.Printf("Loading config from yaml file: %v", configYamlAbsFilepath)
	}
	yfile, err := os.ReadFile(configYamlAbsFilepath)
	Check(err, configYamlAbsFilepath)
	err2 := yaml.Unmarshal(yfile, &conf)
	Check(err2, configYamlAbsFilepath)
	return
}

func ResolveMediaDirs(mediaRootDir string, mediaDirs []string) ([]string, error) {
	rootAbs := ""
	if mediaRootDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to determine current working directory: %w", err)
		}
		rootAbs = cwd
		log.Printf("Using current working directory as media root: %s", rootAbs)
	} else {
		var err error
		rootAbs, err = filepath.Abs(mediaRootDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve mediaRootdir %q: %w", mediaRootDir, err)
		}
		log.Printf("Using mediaRootdir: %s", rootAbs)
	}

	if _, err := os.Stat(rootAbs); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("mediaRootdir does not exist: %s", rootAbs)
		}
		return nil, fmt.Errorf("failed to access mediaRootdir %q: %w", rootAbs, err)
	}

	resolved := make([]string, 0, len(mediaDirs))
	for _, mediaDir := range mediaDirs {
		joined := mediaDir
		if !filepath.IsAbs(mediaDir) {
			joined = filepath.Join(rootAbs, mediaDir)
		} else {
			joined = mediaDir
		}
		resolvedPath, err := filepath.Abs(joined)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve mediaDir %q: %w", mediaDir, err)
		}
		log.Printf("MediaDir %q resolved to %s", mediaDir, resolvedPath)
		if _, err := os.Stat(resolvedPath); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("media dir does not exist: %s", resolvedPath)
			}
			return nil, fmt.Errorf("failed to access media dir %q: %w", resolvedPath, err)
		}
		resolved = append(resolved, resolvedPath)
	}
	return resolved, nil
}
