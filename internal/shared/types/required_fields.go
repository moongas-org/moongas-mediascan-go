package types

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// DecodeRequiredJSON checks field presence before decoding into a type alias.
// Required fields may contain zero values; optional fields need not be present.
func DecodeRequiredJSON(data []byte, target any, fields ...string) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	for _, field := range fields {
		if _, ok := object[field]; !ok {
			return fmt.Errorf("missing required field %q", field)
		}
	}
	return json.Unmarshal(data, target)
}

// DecodeRequiredYAML checks field presence before decoding into a type alias.
func DecodeRequiredYAML(node *yaml.Node, target any, fields ...string) error {
	var object map[string]yaml.Node
	if err := node.Decode(&object); err != nil {
		return err
	}
	for _, field := range fields {
		if _, ok := object[field]; !ok {
			return fmt.Errorf("missing required field %q", field)
		}
	}
	return node.Decode(target)
}
