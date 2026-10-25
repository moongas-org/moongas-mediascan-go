// Package types contains shared data types used across mediascan.
package types

type Date struct {
	Y int `yaml:"y"`
	M int `yaml:"m,omitempty"`
	D int `yaml:"d,omitempty"`
}
