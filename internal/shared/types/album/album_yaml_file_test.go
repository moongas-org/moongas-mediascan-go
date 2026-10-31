package album

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAlbumYamlFileUnmarshal(t *testing.T) {
	data := []byte(`albumData:
  name: Matando güreos
  dateReleased:
    y: 1993
  explicitCover: true
  genres: [Rock, Zydeco]
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
	if !reflect.DeepEqual(a.Genres, []string{"Rock", "Zydeco"}) {
		t.Errorf("unexpected genres: %#v", a.Genres)
	}
	out, err := yaml.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip AlbumYamlFile
	if err := yaml.Unmarshal(out, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundTrip, f) {
		t.Errorf("round trip changed album data: got %#v, want %#v", roundTrip, f)
	}
}
