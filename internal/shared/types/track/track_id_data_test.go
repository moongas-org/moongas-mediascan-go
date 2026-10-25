package track

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTrackIDDataFpCalcDataYaml(t *testing.T) {
	data := []byte(`fpcalcData:
  duration: 123.5
  fingerprint: abc123
`)

	var trackIDData TrackIDData
	if err := yaml.Unmarshal(data, &trackIDData); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if trackIDData.FpCalcData == nil {
		t.Fatal("expected fpcalcData")
	}
	if trackIDData.FpCalcData.Duration == nil || *trackIDData.FpCalcData.Duration != 123.5 {
		t.Errorf("unexpected duration: %#v", trackIDData.FpCalcData.Duration)
	}
	if trackIDData.FpCalcData.Fingerprint == nil || *trackIDData.FpCalcData.Fingerprint != "abc123" {
		t.Errorf("unexpected fingerprint: %#v", trackIDData.FpCalcData.Fingerprint)
	}

	out, err := yaml.Marshal(trackIDData)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{"fpcalcData:", "duration: 123.5", "fingerprint: abc123"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("expected %q in marshaled yaml, got:\n%s", want, out)
		}
	}
}

func TestTrackIDDataAcoustIdDataYaml(t *testing.T) {
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

	var trackIDData TrackIDData
	if err := yaml.Unmarshal(data, &trackIDData); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if trackIDData.AcoustIDData == nil {
		t.Fatal("expected acoustIDData")
	}
	if trackIDData.AcoustIDData.Status != "ok" || len(trackIDData.AcoustIDData.Results) != 1 {
		t.Fatalf("unexpected AcoustID response: %#v", trackIDData.AcoustIDData)
	}
	result := trackIDData.AcoustIDData.Results[0]
	if result.ID != "acoustid-123" || result.Score != 0.98 || len(result.Recordings) != 2 {
		t.Fatalf("unexpected AcoustID result: %#v", result)
	}
	if result.Recordings[0].ID != "recording-456" ||
		len(result.Recordings[0].ISRCs) != 2 ||
		result.Recordings[0].ISRCs[0] != "USSM18500120" ||
		result.Recordings[0].ISRCs[1] != "GBARL8500001" {
		t.Errorf("unexpected first recording: %#v", result.Recordings[0])
	}
	if result.Recordings[1].ID != "recording-789" ||
		len(result.Recordings[1].ISRCs) != 1 ||
		result.Recordings[1].ISRCs[0] != "USSM10203456" {
		t.Errorf("unexpected second recording: %#v", result.Recordings[1])
	}

	out, err := yaml.Marshal(trackIDData)
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
