// Package date defines partial calendar dates used in metadata.
package date

type Date struct {
	Y int `yaml:"y"`
	M int `yaml:"m,omitempty"`
	D int `yaml:"d,omitempty"`
}
