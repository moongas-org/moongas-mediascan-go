package shared

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestArtistDataFileUnmarshal(t *testing.T) {
	data := []byte(`artistData:
  artistNames:
    - Name One
    - Name Two
  city: Athens
  countryCode: GR
  regionCode: I
  languageCodes:
    - el
    - en
  dob:
    y: 1970
  dod:
    y: 2020
    m: 5
    d: 4
`)

	var f ArtistYamlFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	a := f.ArtistData
	if len(a.ArtistNames) != 2 || a.ArtistNames[0] != "Name One" || a.ArtistNames[1] != "Name Two" {
		t.Errorf("unexpected artist names: %#v", a.ArtistNames)
	}
	if a.City != "Athens" || a.CountryCode != "GR" || a.RegionCode != "I" {
		t.Errorf("unexpected location: %q %q %q", a.City, a.CountryCode, a.RegionCode)
	}
	if len(a.LanguageCodes) != 2 || a.LanguageCodes[0] != "el" || a.LanguageCodes[1] != "en" {
		t.Errorf("unexpected language codes: %#v", a.LanguageCodes)
	}
	if a.DOB.Y != 1970 || a.DOB.M != 0 || a.DOB.D != 0 {
		t.Errorf("unexpected dob: %#v", a.DOB)
	}
	if a.DOD.Y != 2020 || a.DOD.M != 5 || a.DOD.D != 4 {
		t.Errorf("unexpected dod: %#v", a.DOD)
	}
}

func TestDateMarshalOmitsZeroMonthAndDay(t *testing.T) {
	out, err := yaml.Marshal(Date{Y: 1970})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(out)
	if !strings.Contains(s, `"y": 1970`) {
		t.Errorf("expected year in marshaled date, got %q", s)
	}
	if strings.Contains(s, `"m":`) || strings.Contains(s, `"d":`) {
		t.Errorf("expected zero month/day to be omitted, got %q", s)
	}
}
