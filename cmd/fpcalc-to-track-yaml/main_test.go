package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"gopkg.in/yaml.v3"
)

func TestRunFpcalcOutputParsing(t *testing.T) {
	var output fpcalcOutput
	if err := json.Unmarshal([]byte(`{"duration":12.5,"fingerprint":"abc"}`), &output); err != nil {
		t.Fatal(err)
	}
	if output.Duration == nil || *output.Duration != 12.5 || output.Fingerprint == nil || *output.Fingerprint != "abc" {
		t.Fatalf("unexpected fpcalc output: %#v", output)
	}
}

func TestNeedsFpcalcData(t *testing.T) {
	for _, tt := range []struct {
		name string
		yaml string
		want bool
	}{
		{
			name: "missing track identifiers",
			yaml: "trackData:\n  title: Song\n",
			want: true,
		},
		{
			name: "missing fpcalc data",
			yaml: "trackData:\n  title: Song\ntrackIdData:\n  - isrc: ABC\n",
			want: true,
		},
		{
			name: "missing fingerprint",
			yaml: "trackIdData:\n  - fpcalcData:\n      duration: 12.5\n",
			want: true,
		},
		{
			name: "complete fpcalc data",
			yaml: "trackIdData:\n  - fpcalcData:\n      duration: 12.5\n      fingerprint: abc\n",
			want: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := needsFpcalcData([]byte(tt.yaml))
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("needsFpcalcData = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestUpdateTrackYamlFilePreservesOtherDataAndBacksUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.yml")
	original := []byte(`trackData:
  title: Song
trackIdData:
  - isrc: ABC
    spotifyId: track-id
`)
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatal(err)
	}

	duration := 12.5
	fingerprint := "new-fingerprint"
	updated, err := updateTrackYamlFile(path, original, track.TrackFpCalcData{
		Duration:    &duration,
		Fingerprint: &fingerprint,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected YAML to be updated")
	}

	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != string(original) {
		t.Fatalf("backup differs from original: %q", backup)
	}
	updatedYaml, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedYaml), "spotifyId: track-id") ||
		!strings.Contains(string(updatedYaml), "isrc: ABC") ||
		!strings.Contains(string(updatedYaml), "title: Song") {
		t.Fatalf("unrelated YAML fields were not preserved:\n%s", updatedYaml)
	}
	needsUpdate, err := needsFpcalcData(updatedYaml)
	if err != nil {
		t.Fatal(err)
	}
	if needsUpdate {
		t.Fatal("updated YAML is still missing fpcalc data")
	}
}

func TestUpdateTrackYamlFileUpdatesExistingFpcalcData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.yml")
	original := []byte(`trackData:
  title: Song
trackIdData:
  - isrc: ABC
    fpcalcData:
      duration: 1
      fingerprint: old
`)
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	duration := 2.25
	fingerprint := "replacement"
	updated, err := updateTrackYamlFile(path, original, track.TrackFpCalcData{
		Duration:    &duration,
		Fingerprint: &fingerprint,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected fpcalc data to be replaced")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	trackIDs := root["trackIdData"].([]interface{})
	entry := trackIDs[0].(map[string]interface{})
	fpcalcData := entry["fpcalcData"].(map[string]interface{})
	if fpcalcData["duration"] != 2.25 || fpcalcData["fingerprint"] != "replacement" || entry["isrc"] != "ABC" {
		t.Fatalf("unexpected updated YAML: %#v", root)
	}
}

func TestUpdateTrackYamlFileDoesNotWriteWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.yml")
	original := []byte("trackData:\n  title: Song\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	duration := 12.5
	fingerprint := "abc"
	result := track.TrackFpCalcData{Duration: &duration, Fingerprint: &fingerprint}
	if _, err := updateTrackYamlFile(path, original, result); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup to be created: %v", err)
	}
	if changed, err := updateTrackYamlFile(path, updated, result); err != nil {
		t.Fatal(err)
	} else if changed {
		t.Fatal("expected identical fpcalc data not to modify YAML")
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected existing backup to remain: %v", err)
	}
}
