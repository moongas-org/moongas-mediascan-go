package shared

type Date struct {
	Y int `yaml:"y"` // Mandatory or non-pointer field
	// Optional fields use pointers and 'omitempty' for clean re-marshaling
	M int `yaml:"m,omitempty"`
	D int `yaml:"d,omitempty"`
}

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
	DOB Date `yaml:"dob,omitempty"`
	DOD Date `yaml:"dod,omitempty"`
}
