package album

import (
	"reflect"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
	"gopkg.in/yaml.v3"
)

func TestAlbumYamlRoundTrip(t *testing.T) {
	want := Album{
		AlbumYamlFile: AlbumYamlFile{
			AlbumData: AlbumData{
				Name:          "Example Album",
				DateReleased:  date.Date{Y: 1993, M: 5, D: 4},
				ExplicitCover: true,
			},
		},
		Path: "/music/Artist/Album",
	}
	out, err := yaml.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Album
	if err := yaml.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("album directory = %+v, want %+v", got, want)
	}
}
