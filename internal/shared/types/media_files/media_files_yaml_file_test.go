package media_files

import (
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"gopkg.in/yaml.v3"
)

// The files.yml output is consumed by external tooling, so the top-level
// key and field names must remain stable.
func TestMediaFilesYamlFileMarshalKeys(t *testing.T) {
	f := MediaFilesYamlFile{
		Files: []media_file.MediaFileData{
			{Path: "/music/Artist/Album/track.mp3", Title: "Song", Year: 1999},
		},
	}

	out, err := yaml.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(out)
	for _, want := range []string{"files:", "path: /music/Artist/Album/track.mp3", "title: Song", "year: 1999"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in marshaled yaml, got:\n%s", want, s)
		}
	}
}
