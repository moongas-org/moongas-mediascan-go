package artist

import (
	"reflect"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
	"gopkg.in/yaml.v3"
)

func TestArtistYamlRoundTrip(t *testing.T) {
	want := Artist{
		ArtistYamlFile: ArtistYamlFile{
			ArtistData: ArtistData{
				ArtistNames:   []string{"Example Artist"},
				City:          "Liverpool",
				CountryCode:   "GB",
				RegionCode:    "ENG",
				LanguageCodes: []string{"en"},
				DOB:           date.Date{Y: 1960, M: 10, D: 9},
			},
		},
		Path: "/music/Artist",
	}
	out, err := yaml.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Artist
	if err := yaml.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("artist directory = %+v, want %+v", got, want)
	}
}
