package album

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAlbumYamlFileUnmarshal(t *testing.T) {
	data := []byte(`albumData:
  name: Matando güreos
  dateReleased:
    y: 1993
  explicitCover: true
`)

	var f AlbumYamlFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	a := f.AlbumData
	if a.Name != "Matando güreos" {
		t.Errorf("unexpected album name: %q", a.Name)
	}
	if a.DateReleased.Y != 1993 || a.DateReleased.M != 0 || a.DateReleased.D != 0 {
		t.Errorf("unexpected release date: %#v", a.DateReleased)
	}
	if !a.ExplicitCover {
		t.Error("expected explicit cover to be true")
	}
}
