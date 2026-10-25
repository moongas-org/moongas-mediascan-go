package track

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTrackIDDataFpcalcDataYaml(t *testing.T) {
	data := []byte(`fpcalcData:
  duration: 123.5
  fingerprint: abc123
`)

	var trackIDData TrackIDData
	if err := yaml.Unmarshal(data, &trackIDData); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if trackIDData.FpcalcData == nil {
		t.Fatal("expected fpcalcData")
	}
	if trackIDData.FpcalcData.Duration == nil || *trackIDData.FpcalcData.Duration != 123.5 {
		t.Errorf("unexpected duration: %#v", trackIDData.FpcalcData.Duration)
	}
	if trackIDData.FpcalcData.Fingerprint == nil || *trackIDData.FpcalcData.Fingerprint != "abc123" {
		t.Errorf("unexpected fingerprint: %#v", trackIDData.FpcalcData.Fingerprint)
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
