package shared

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The artists.yml output is consumed by external tooling, so the top-level
// key and field names must remain stable.
func TestArtistsYamlFileMarshalKeys(t *testing.T) {
	f := ArtistsYamlFile{
		Artists: []ArtistsYamlArtistData{
			{
				Path: "/music/Beatles",
				ArtistData: ArtistData{
					ArtistNames:   []string{"The Beatles"},
					City:          "Liverpool",
					CountryCode:   "GB",
					RegionCode:    "ENG",
					LanguageCodes: []string{"en"},
				},
			},
		},
	}

	out, err := yaml.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(out)
	for _, want := range []string{"artists:", "path: /music/Beatles", "artistData:", "- The Beatles"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in marshaled yaml, got:\n%s", want, s)
		}
	}
}
