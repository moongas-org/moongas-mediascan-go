// Package main updates fpcalc metadata in existing track YAML files.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	media_file_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/media_files"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"gopkg.in/yaml.v3"
)

type fpcalcOutput struct {
	Duration    *float64 `json:"duration"`
	Fingerprint *string  `json:"fingerprint"`
}

func main() {
	cli.LogCurrentDir()

	overwriteMode, args, err := cli.ParseCommandArgs(os.Args, 1, 2)
	if err != nil {
		log.Println("Error:", err)
		log.Println("Usage: go run cmd/fpcalc-to-track-yaml <config-yaml> [media-root] [--overwrite-existing|--overwrite-newer]")
		log.Println("Example: go run cmd/fpcalc-to-track-yaml mediascan-config.yml /path/to/root --overwrite-existing")
		os.Exit(1)
	}
	mediaRootDir := ""
	if len(args) == 2 {
		mediaRootDir = args[1]
	}

	conf, err := cli.ResolveScanConfig(args[0], mediaRootDir)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	files := media_file_scanner.ScanFiles(conf)
	countUpdated := 0
	countSkipped := 0
	countFailed := 0

	for _, mediaFile := range files {
		trackYamlPath := trackYamlPath(mediaFile)
		yamlData, err := os.ReadFile(trackYamlPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Skipping %s (no associated track YAML)", mediaFile.Path)
				countSkipped++
				continue
			}
			log.Printf("ERROR reading %s: %v", trackYamlPath, err)
			countFailed++
			continue
		}

		needsFpcalc, err := needsFpcalcData(yamlData)
		if err != nil {
			log.Printf("ERROR reading fpcalc data from %s: %v", trackYamlPath, err)
			countFailed++
			continue
		}
		if !needsFpcalc && overwriteMode != cli.SkipExisting {
			needsFpcalc, err = cli.ShouldOverwriteFile(trackYamlPath, mediaFile.Path, overwriteMode)
			if err != nil {
				log.Printf("ERROR checking overwrite for %s: %v", trackYamlPath, err)
				countFailed++
				continue
			}
		}
		if !needsFpcalc {
			log.Printf("Skipping %s (fpcalc data already exists)", trackYamlPath)
			countSkipped++
			continue
		}

		fpcalcData, err := runFpcalc(mediaFile.Path)
		if err != nil {
			log.Printf("ERROR running fpcalc for %s: %v", mediaFile.Path, err)
			countFailed++
			continue
		}
		updated, err := updateTrackYamlFile(trackYamlPath, yamlData, fpcalcData)
		if err != nil {
			log.Printf("ERROR updating %s: %v", trackYamlPath, err)
			countFailed++
			continue
		}
		if updated {
			log.Printf("Updated %s", trackYamlPath)
			countUpdated++
		} else {
			countSkipped++
		}
	}

	log.Printf("Scanned %d media files", len(files))
	log.Printf("Updated %d track YAML (.yml) files", countUpdated)
	log.Printf("Skipped %d media files", countSkipped)
	if countFailed > 0 {
		log.Printf("Failed to update %d track YAML (.yml) files", countFailed)
		os.Exit(1)
	}
}

func trackYamlPath(mediaFile media_file.MediaFileData) string {
	ext := filepath.Ext(mediaFile.Path)
	return filepath.Join(mediaFile.AlbumPath, strings.TrimSuffix(filepath.Base(mediaFile.Path), ext)+".yml")
}

func runFpcalc(mediaPath string) (track.TrackFpCalcData, error) {
	output, err := exec.Command("fpcalc", "-json", mediaPath).CombinedOutput()
	if err != nil {
		return track.TrackFpCalcData{}, fmt.Errorf("fpcalc -json failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	var result fpcalcOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return track.TrackFpCalcData{}, fmt.Errorf("parse fpcalc JSON: %w", err)
	}
	if result.Duration == nil || result.Fingerprint == nil || *result.Fingerprint == "" {
		return track.TrackFpCalcData{}, fmt.Errorf("fpcalc JSON is missing duration or fingerprint")
	}

	return track.TrackFpCalcData{
		Duration:    result.Duration,
		Fingerprint: result.Fingerprint,
	}, nil
}

func needsFpcalcData(data []byte) (bool, error) {
	root, err := parseYamlRoot(data)
	if err != nil {
		return false, err
	}
	mediaFiles, err := mediaFileSequence(root, false)
	if err != nil {
		return false, err
	}
	if mediaFiles == nil {
		return true, nil
	}
	for _, entry := range mediaFiles.Content {
		if entry.Kind != yaml.MappingNode {
			return false, fmt.Errorf("mediaFileData entries must be mappings")
		}
		fpcalcNode := mappingValue(entry, "fpcalcData")
		if fpcalcNode != nil {
			return !validFpcalcNode(fpcalcNode), nil
		}
	}
	return true, nil
}

func updateTrackYamlFile(path string, original []byte, data track.TrackFpCalcData) (bool, error) {
	if data.Duration == nil || data.Fingerprint == nil || *data.Fingerprint == "" {
		return false, fmt.Errorf("fpcalc data must include duration and fingerprint")
	}
	root, err := parseYamlRoot(original)
	if err != nil {
		return false, err
	}
	mediaFiles, err := mediaFileSequence(root, true)
	if err != nil {
		return false, err
	}

	var fpcalcNode *yaml.Node
	for _, entry := range mediaFiles.Content {
		if entry.Kind != yaml.MappingNode {
			return false, fmt.Errorf("mediaFileData entries must be mappings")
		}
		if fpcalcNode = mappingValue(entry, "fpcalcData"); fpcalcNode != nil {
			break
		}
	}
	if fpcalcNode == nil {
		var entry *yaml.Node
		if len(mediaFiles.Content) > 0 {
			entry = mediaFiles.Content[0]
		} else {
			entry = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			mediaFiles.Content = append(mediaFiles.Content, entry)
		}
		appendMappingValue(entry, "fpcalcData", &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"})
		fpcalcNode = mappingValue(entry, "fpcalcData")
	} else if fpcalcNode.Kind != yaml.MappingNode {
		if fpcalcNode.Tag != "!!null" {
			return false, fmt.Errorf("fpcalcData must be a mapping")
		}
		*fpcalcNode = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}

	duration := strconv.FormatFloat(*data.Duration, 'f', -1, 64)
	changed := setMappingScalar(fpcalcNode, "duration", duration, "!!float")
	changed = setMappingScalar(fpcalcNode, "fingerprint", *data.Fingerprint, "!!str") || changed
	if !changed {
		return false, nil
	}

	updatedYaml, err := yaml.Marshal(root)
	if err != nil {
		return false, fmt.Errorf("marshal updated YAML: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.WriteFile(path+".bak", original, info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("write backup %s.bak: %w", path, err)
	}
	if err := os.WriteFile(path, updatedYaml, info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("write updated YAML %s: %w", path, err)
	}
	return true, nil
}

func parseYamlRoot(data []byte) (*yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("track YAML root must be a mapping")
	}
	return document.Content[0], nil
}

func mediaFileSequence(root *yaml.Node, create bool) (*yaml.Node, error) {
	mediaFiles := mappingValue(root, "mediaFileData")
	if mediaFiles == nil {
		if !create {
			return nil, nil
		}
		mediaFiles = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		appendMappingValue(root, "mediaFileData", mediaFiles)
		return mediaFiles, nil
	}
	if mediaFiles.Tag == "!!null" {
		if !create {
			return nil, nil
		}
		*mediaFiles = yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	}
	if mediaFiles.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("mediaFileData must be a sequence")
	}
	return mediaFiles, nil
}

func validFpcalcNode(node *yaml.Node) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}
	durationNode := mappingValue(node, "duration")
	fingerprintNode := mappingValue(node, "fingerprint")
	if durationNode == nil || durationNode.Tag == "!!null" || fingerprintNode == nil || fingerprintNode.Tag == "!!null" {
		return false
	}
	var duration float64
	var fingerprint string
	if durationNode.Decode(&duration) != nil || fingerprintNode.Decode(&fingerprint) != nil {
		return false
	}
	return fingerprint != ""
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func appendMappingValue(mapping *yaml.Node, key string, value *yaml.Node) {
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		value,
	)
}

func setMappingScalar(mapping *yaml.Node, key, value, tag string) bool {
	node := mappingValue(mapping, key)
	if node == nil {
		appendMappingValue(mapping, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value})
		return true
	}
	if node.Kind == yaml.ScalarNode && node.Tag == tag && node.Value == value {
		return false
	}
	comment := node.LineComment
	headComment := node.HeadComment
	footComment := node.FootComment
	*node = yaml.Node{
		Kind:        yaml.ScalarNode,
		Tag:         tag,
		Value:       value,
		Style:       node.Style,
		HeadComment: headComment,
		LineComment: comment,
		FootComment: footComment,
	}
	return true
}
