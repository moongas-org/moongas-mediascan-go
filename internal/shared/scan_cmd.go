package shared

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type OverwriteMode int

const (
	SkipExisting OverwriteMode = iota
	OverwriteExisting
	OverwriteNewer
)

func LogCurrentDir() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to determine current working directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", cwd)
	}
}

func ParseCommandArgs(argv []string, minArgs int, maxArgs int) (OverwriteMode, []string, error) {
	mode := SkipExisting
	args := make([]string, 0, len(argv)-1)
	for _, arg := range argv[1:] {
		switch arg {
		case "--overwrite-existing":
			if mode != SkipExisting {
				return SkipExisting, nil, fmt.Errorf("--overwrite-existing and --overwrite-newer are mutually exclusive")
			}
			mode = OverwriteExisting
		case "--overwrite-newer":
			if mode != SkipExisting {
				return SkipExisting, nil, fmt.Errorf("--overwrite-existing and --overwrite-newer are mutually exclusive")
			}
			mode = OverwriteNewer
		default:
			args = append(args, arg)
		}
	}
	if len(args) < minArgs || len(args) > maxArgs {
		return SkipExisting, nil, fmt.Errorf("invalid arguments")
	}
	return mode, args, nil
}

func ResolveScanConfig(configYamlFilepath string, mediaRootDir string) (MediascanConf, error) {
	conf := LoadConf(configYamlFilepath)
	resolvedMediaDirs, err := ResolveMediaDirs(mediaRootDir, conf.MediaDirs)
	if err != nil {
		return conf, err
	}
	conf.MediaDirs = resolvedMediaDirs
	return conf, nil
}

func ResolveOutputAbsPath(outputPath string) (string, error) {
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path to output file: %w", err)
	}
	return absPath, nil
}

func ShouldOverwriteFile(destPath string, sourcePath string, mode OverwriteMode) (bool, error) {
	if mode == SkipExisting {
		_, err := os.Stat(destPath)
		if err == nil {
			return false, nil
		}
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	_, err := os.Stat(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	if mode == OverwriteNewer {
		return true, nil
	}

	if sourcePath == "" {
		return true, nil
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	destInfo, err := os.Stat(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	return destInfo.ModTime().Before(sourceInfo.ModTime()), nil
}

func LatestSourcePath(paths ...string) string {
	var newest string
	var newestTime time.Time
	for _, p := range paths {
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if newest == "" || info.ModTime().After(newestTime) {
			newest = p
			newestTime = info.ModTime()
		}
	}
	return newest
}
