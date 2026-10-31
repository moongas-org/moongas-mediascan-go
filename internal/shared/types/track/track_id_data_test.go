package track

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track/streaming"
	"gopkg.in/yaml.v3"
)

func TestTrackIDDataStreamingDataYaml(t *testing.T) {
	input := []byte(`streamingData:
  - streamingService: spotify
    isrc: USSM10027739
    trackUrl: https://open.spotify.com/track/0ZYf6N5eBeXStMri0mf3DZ
    trackId: 0ZYf6N5eBeXStMri0mf3DZ
    apiData:
      tracks:
        items:
          - id: 0ZYf6N5eBeXStMri0mf3DZ
            duration_ms: 263573
  - streamingService: apple-music
    trackUrl: https://music.apple.com/track/example
    trackId: example
    apiData: null
`)
	var data TrackIDData
	if err := yaml.Unmarshal(input, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.StreamingData) != 2 {
		t.Fatalf("expected two streaming entries, got %+v", data.StreamingData)
	}
	first := data.StreamingData[0]
	if first.StreamingService != streaming.Spotify ||
		first.ISRC == nil || *first.ISRC != "USSM10027739" ||
		first.TrackID != "0ZYf6N5eBeXStMri0mf3DZ" ||
		first.TrackURL != "https://open.spotify.com/track/0ZYf6N5eBeXStMri0mf3DZ" {
		t.Errorf("unexpected first entry: %+v", first)
	}
	if _, ok := first.APIData.(map[string]any); !ok {
		t.Fatalf("expected generic API data map, got %T", first.APIData)
	}
	if data.StreamingData[1].StreamingService != streaming.AppleMusic ||
		data.StreamingData[1].TrackID != "example" || data.StreamingData[1].ISRC != nil ||
		data.StreamingData[1].APIData != nil {
		t.Errorf("unexpected second entry: %+v", data.StreamingData[1])
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var want, got map[string]any
	if err := yaml.Unmarshal(input, &want); err != nil {
		t.Fatalf("unmarshal expected YAML: %v", err)
	}
	if err := yaml.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal output YAML: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("streaming data changed:\ngot: %s\nwant: %s", out, input)
	}
}

func TestTrackIDDataStreamingDataJSON(t *testing.T) {
	input := []byte(`{"streamingData":[{"streamingService":"spotify","trackUrl":"https://open.spotify.com/track/example","trackId":"example","apiData":{"tracks":{"items":[]}}}]}`)
	var data TrackIDData
	if err := json.Unmarshal(input, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.StreamingData) != 1 || data.StreamingData[0].TrackID != "example" {
		t.Fatalf("unexpected streaming data: %+v", data.StreamingData)
	}
	out, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(out, &fields); err != nil {
		t.Fatalf("unmarshal fields: %v", err)
	}
	if _, ok := fields["streamingData"]; !ok {
		t.Fatalf("missing streamingData field in %s", out)
	}
}

func TestTrackIDDataOmitsEmptyStreamingData(t *testing.T) {
	for _, entries := range [][]streaming.StreamingData[any]{nil, {}} {
		data := TrackIDData{StreamingData: entries}
		out, err := yaml.Marshal(data)
		if err != nil {
			t.Fatalf("marshal YAML: %v", err)
		}
		if strings.Contains(string(out), "streamingData") {
			t.Errorf("expected empty streamingData to be omitted: %s", out)
		}
		out, err = json.Marshal(data)
		if err != nil {
			t.Fatalf("marshal JSON: %v", err)
		}
		if strings.Contains(string(out), "streamingData") {
			t.Errorf("expected empty streamingData to be omitted: %s", out)
		}
	}
}

func TestTrackMediaFileDataFpCalcDataYaml(t *testing.T) {
	data := []byte(`fpcalcData:
  duration: 123.5
  fingerprint: abc123
`)

	var mediaFileData TrackMediaFileData
	if err := yaml.Unmarshal(data, &mediaFileData); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if mediaFileData.FpCalcData == nil {
		t.Fatal("expected fpcalcData")
	}
	if mediaFileData.FpCalcData.Duration == nil || *mediaFileData.FpCalcData.Duration != 123.5 {
		t.Errorf("unexpected duration: %#v", mediaFileData.FpCalcData.Duration)
	}
	if mediaFileData.FpCalcData.Fingerprint == nil || *mediaFileData.FpCalcData.Fingerprint != "abc123" {
		t.Errorf("unexpected fingerprint: %#v", mediaFileData.FpCalcData.Fingerprint)
	}

	out, err := yaml.Marshal(mediaFileData)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{"fpcalcData:", "duration: 123.5", "fingerprint: abc123"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("expected %q in marshaled yaml, got:\n%s", want, out)
		}
	}
}

func TestTrackMediaFileDataAcoustIdDataYaml(t *testing.T) {
	data := []byte(`acoustidData:
  results:
    - id: acoustid-123
      recordings:
        - id: recording-456
          isrcs:
            - USSM18500120
            - GBARL8500001
        - id: recording-789
          isrcs:
            - USSM10203456
      score: 0.98
  status: ok
`)

	var mediaFileData TrackMediaFileData
	if err := yaml.Unmarshal(data, &mediaFileData); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if mediaFileData.AcoustIDData == nil {
		t.Fatal("expected acoustIDData")
	}
	if mediaFileData.AcoustIDData.Status != "ok" || len(mediaFileData.AcoustIDData.Results) != 1 {
		t.Fatalf("unexpected AcoustID response: %#v", mediaFileData.AcoustIDData)
	}
	result := mediaFileData.AcoustIDData.Results[0]
	if result.ID != "acoustid-123" || result.Score != 0.98 || len(result.Recordings) != 2 {
		t.Fatalf("unexpected AcoustID result: %#v", result)
	}
	if result.Recordings[0].ID != "recording-456" ||
		result.Recordings[0].ISRCs == nil ||
		len(*result.Recordings[0].ISRCs) != 2 ||
		(*result.Recordings[0].ISRCs)[0] != "USSM18500120" ||
		(*result.Recordings[0].ISRCs)[1] != "GBARL8500001" {
		t.Errorf("unexpected first recording: %#v", result.Recordings[0])
	}
	if result.Recordings[1].ID != "recording-789" ||
		result.Recordings[1].ISRCs == nil ||
		len(*result.Recordings[1].ISRCs) != 1 ||
		(*result.Recordings[1].ISRCs)[0] != "USSM10203456" {
		t.Errorf("unexpected second recording: %#v", result.Recordings[1])
	}

	out, err := yaml.Marshal(mediaFileData)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{
		"acoustidData:",
		"status: ok",
		"id: acoustid-123",
		"id: recording-456",
		"USSM18500120",
		"GBARL8500001",
		"id: recording-789",
		"USSM10203456",
		"score: 0.98",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("expected %q in marshaled yaml, got:\n%s", want, out)
		}
	}
}
