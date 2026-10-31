// Package artist defines artist metadata types and YAML representations.
package artist

import "github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"

// ArtistData contains the contents of the artist_data block of an artist.yml file.
// keep me in sync with moongas-mediascan-python/src/mediascan/artist_data.py
type ArtistData struct {
	// Mandatory or non-pointer fields:
	ArtistNames   []string `yaml:"artistNames"`
	City          string   `yaml:"city"`
	CountryCode   string   `yaml:"countryCode"`
	RegionCode    string   `yaml:"regionCode"`
	LanguageCodes []string `yaml:"languageCodes"`
	// Optional fields use pointers and 'omitempty' for clean re-marshaling
	DOB date.Date `yaml:"dob,omitempty"`
	DOD date.Date `yaml:"dod,omitempty"`
}
