package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestScanToTrackYAMLQuotesBooleanStrings(t *testing.T) {
	if os.Getenv("SCAN_TO_TRACK_YAML_TEST_HELPER") == "1" {
		os.Args = []string{"scan-to-track-yaml", os.Getenv("SCAN_TO_TRACK_YAML_TEST_CONFIG")}
		main()
		return
	}

	values := []string{
		"y", "Y", "yes", "Yes", "YES", "true", "True", "TRUE", "on", "On", "ON",
		"n", "N", "no", "No", "NO", "false", "False", "FALSE", "off", "Off", "OFF",
		"Yesterday",
	}
	dir := t.TempDir()
	mediaDir := filepath.Join(dir, "media")
	if err := os.Mkdir(mediaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i, value := range values {
		path := filepath.Join(mediaDir, fmt.Sprintf("%02d.mp3", i))
		if err := os.WriteFile(path, testID3Tag(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	config, err := yaml.Marshal(map[string]any{
		"mediaDirs": []string{mediaDir},
		"mediaExts": []string{".mp3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestScanToTrackYAMLQuotesBooleanStrings$")
	cmd.Env = append(os.Environ(), "SCAN_TO_TRACK_YAML_TEST_HELPER=1", "SCAN_TO_TRACK_YAML_TEST_CONFIG="+configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("scan-to-track-yaml: %v\n%s", err, output)
	}

	for i, value := range values {
		t.Run(value, func(t *testing.T) {
			path := filepath.Join(mediaDir, fmt.Sprintf("%02d.yml", i))
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var doc yaml.Node
			if err := yaml.Unmarshal(data, &doc); err != nil {
				t.Fatal(err)
			}
			root := doc.Content[0]
			if len(root.Content) != 2 || root.Content[0].Value != "trackData" {
				t.Fatalf("unexpected YAML structure: %s", data)
			}
			fields := root.Content[1]
			if len(fields.Content) != 12 {
				t.Fatalf("unexpected trackData fields: %s", data)
			}
			expected := map[string]bool{
				"title": true, "artist": true, "albumartist": true,
				"album": true, "genre": true, "year": true,
			}
			for j := 0; j < len(fields.Content); j += 2 {
				name, node := fields.Content[j].Value, fields.Content[j+1]
				if !expected[name] {
					t.Errorf("unexpected or repeated field %q", name)
				}
				delete(expected, name)
				if name == "year" {
					if node.Tag != "!!int" || node.Value != "2026" {
						t.Errorf("year changed type or value: %s", data)
					}
					continue
				}
				if node.Tag != "!!str" || node.Value != value {
					t.Errorf("%s: got tag %s value %q, want string %q", name, node.Tag, node.Value, value)
				}
				quoted := node.Style&(yaml.SingleQuotedStyle|yaml.DoubleQuotedStyle) != 0
				if value != "Yesterday" && !quoted {
					t.Errorf("%s: %q is not quoted in %s", name, value, data)
				}
				if value == "Yesterday" && quoted {
					t.Errorf("%s: ordinary value unexpectedly quoted in %s", name, data)
				}
			}
			if len(expected) != 0 {
				t.Errorf("missing fields: %v", expected)
			}
		})
	}
}

func testID3Tag(value string) []byte {
	var body []byte
	for _, frame := range [][2]string{
		{"TIT2", value}, {"TPE1", value}, {"TPE2", value},
		{"TALB", value}, {"TCON", value}, {"TYER", "2026"},
	} {
		data := append([]byte{0}, frame[1]...)
		size := len(data)
		body = append(body, frame[0]...)
		body = append(body, byte(size>>24), byte(size>>16), byte(size>>8), byte(size), 0, 0)
		body = append(body, data...)
	}
	size := len(body)
	header := []byte{'I', 'D', '3', 3, 0, 0, byte(size >> 21 & 0x7f), byte(size >> 14 & 0x7f), byte(size >> 7 & 0x7f), byte(size & 0x7f)}
	return append(header, body...)
}
