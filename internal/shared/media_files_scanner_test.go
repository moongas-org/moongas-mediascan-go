package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFilesFiltersMediaFiles(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("create album directory: %v", err)
	}

	trackPath := filepath.Join(albumDir, "Track.mp3")
	if err := os.WriteFile(trackPath, []byte("not a complete mp3"), 0o644); err != nil {
		t.Fatalf("create track: %v", err)
	}
	if err := os.WriteFile(filepath.Join(albumDir, "cover.jpg"), []byte("cover"), 0o644); err != nil {
		t.Fatalf("create cover: %v", err)
	}
	excludedPath := filepath.Join(albumDir, "Excluded.mp3")
	if err := os.WriteFile(excludedPath, []byte("not a complete mp3"), 0o644); err != nil {
		t.Fatalf("create excluded track: %v", err)
	}

	files := ScanFiles(MediascanConf{
		MediaDirs:    []string{root},
		MediaExts:    []string{".mp3"},
		ExcludePaths: []string{"Excluded.mp3"},
	})

	if len(files) != 1 {
		t.Fatalf("expected 1 media file, got %d", len(files))
	}
	if files[0].Path != trackPath {
		t.Errorf("got path %q, want %q", files[0].Path, trackPath)
	}
	if files[0].Title != "Track" {
		t.Errorf("got title %q, want %q", files[0].Title, "Track")
	}
}

// fakeMp3 builds a file whose only content is a minimal ID3v2.3 tag holding
// the given text frames, e.g. fakeMp3([2]string{"TPE1", "The Artist"}).
// No audio frames are included; tag.ReadFrom can still read the metadata.
func fakeMp3(frames ...[2]string) []byte {
	var body []byte
	for _, frame := range frames {
		data := append([]byte{0x00}, []byte(frame[1])...) // 0x00 = ISO-8859-1 text encoding
		size := len(data)
		header := []byte(frame[0])
		header = append(header, byte(size>>24), byte(size>>16), byte(size>>8), byte(size)) // v2.3 frame size (big-endian)
		header = append(header, 0x00, 0x00)                                                // frame flags
		body = append(body, header...)
		body = append(body, data...)
	}
	n := len(body)
	tagHeader := []byte{'I', 'D', '3', 0x03, 0x00, 0x00} // ID3v2.3.0, no flags
	tagHeader = append(tagHeader,
		byte(n>>21&0x7F), byte(n>>14&0x7F), byte(n>>7&0x7F), byte(n&0x7F)) // syncsafe tag size
	return append(tagHeader, body...)
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func TestScanFiles_ReadsTags(t *testing.T) {
	root := t.TempDir()
	trackPath := filepath.Join(root, "Artist", "Album", "track.mp3")
	writeFile(t, trackPath, fakeMp3(
		[2]string{"TIT2", "Song Title"},
		[2]string{"TPE1", "The Artist"},
		[2]string{"TPE2", "The Album Artist"},
		[2]string{"TALB", "The Album"},
		[2]string{"TCON", "Rock"},
		[2]string{"TYER", "1999"},
	))

	files := ScanFiles(MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3"}})

	if len(files) != 1 {
		t.Fatalf("expected 1 media file, got %d", len(files))
	}
	m := files[0]
	if m.Title != "Song Title" {
		t.Errorf("got title %q, want %q", m.Title, "Song Title")
	}
	if m.Artist != "The Artist" {
		t.Errorf("got artist %q, want %q", m.Artist, "The Artist")
	}
	if m.AlbumArtist != "The Album Artist" {
		t.Errorf("got album artist %q, want %q", m.AlbumArtist, "The Album Artist")
	}
	if m.Album != "The Album" {
		t.Errorf("got album %q, want %q", m.Album, "The Album")
	}
	if m.Genre != "Rock" {
		t.Errorf("got genre %q, want %q", m.Genre, "Rock")
	}
	if m.Year != 1999 {
		t.Errorf("got year %d, want %d", m.Year, 1999)
	}
	if m.Format == "" {
		t.Error("expected format to be set from tags")
	}
}

func TestScanFiles_PopulatesPathDerivedMetadata(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "Artist", "Album")
	trackPath := filepath.Join(albumDir, "01 Song.mp3")
	content := []byte("not a complete mp3") // unreadable tags: filename fallbacks apply
	writeFile(t, trackPath, content)

	files := ScanFiles(MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3"}})

	if len(files) != 1 {
		t.Fatalf("expected 1 media file, got %d", len(files))
	}
	m := files[0]
	if m.AlbumPath != albumDir {
		t.Errorf("got album path %q, want %q", m.AlbumPath, albumDir)
	}
	if want := filepath.Join(root, "Artist"); m.ArtistPath != want {
		t.Errorf("got artist path %q, want %q", m.ArtistPath, want)
	}
	if m.Size != int64(len(content)) {
		t.Errorf("got size %d, want %d", m.Size, len(content))
	}
	if m.Title != "01 Song" {
		t.Errorf("got title %q, want filename fallback %q", m.Title, "01 Song")
	}
	if m.Format != "" {
		t.Errorf("expected empty format for unreadable tags, got %q", m.Format)
	}
	if m.Year != 0 {
		t.Errorf("expected zero year for unreadable tags, got %d", m.Year)
	}
	if m.ModTime.IsZero() {
		t.Error("expected mod time to be set")
	}
}

func TestScanFiles_SortsResults(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    string
		wantOrder []string
	}{
		{name: "by year", sortBy: "year", wantOrder: []string{"b.mp3", "a.mp3", "c.mp3"}},
		{name: "by artist", sortBy: "artist", wantOrder: []string{"a.mp3", "c.mp3", "b.mp3"}},
		{name: "no sort keeps walk order", sortBy: "", wantOrder: []string{"a.mp3", "b.mp3", "c.mp3"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			albumDir := filepath.Join(root, "Artist", "Album")
			writeFile(t, filepath.Join(albumDir, "a.mp3"), fakeMp3([2]string{"TYER", "2001"}, [2]string{"TPE1", "Alpha"}))
			writeFile(t, filepath.Join(albumDir, "b.mp3"), fakeMp3([2]string{"TYER", "1999"}, [2]string{"TPE1", "Gamma"}))
			writeFile(t, filepath.Join(albumDir, "c.mp3"), fakeMp3([2]string{"TYER", "2005"}, [2]string{"TPE1", "Beta"}))

			files := ScanFiles(MediascanConf{
				MediaDirs: []string{root},
				MediaExts: []string{".mp3"},
				SortBy:    test.sortBy,
			})

			if len(files) != len(test.wantOrder) {
				t.Fatalf("expected %d media files, got %d", len(test.wantOrder), len(files))
			}
			for i, want := range test.wantOrder {
				if got := filepath.Base(files[i].Path); got != want {
					t.Errorf("file %d: got %q, want %q", i, got, want)
				}
			}
		})
	}
}

func TestScanFiles_ExcludesByTags(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "Artist", "Album")
	keepPath := filepath.Join(albumDir, "keep.mp3")
	writeFile(t, keepPath, fakeMp3(
		[2]string{"TIT2", "Keep Title"},
		[2]string{"TPE1", "Keep Artist"},
		[2]string{"TPE2", "Keep Album Artist"},
		[2]string{"TALB", "Keep Album"},
		[2]string{"TCON", "Keep Genre"},
	))
	writeFile(t, filepath.Join(albumDir, "live.mp3"), fakeMp3([2]string{"TIT2", "Keep Title (Live)"}))
	writeFile(t, filepath.Join(albumDir, "badartist.mp3"), fakeMp3([2]string{"TPE1", "Bad Artist"}))
	writeFile(t, filepath.Join(albumDir, "badalbum.mp3"), fakeMp3([2]string{"TALB", "Bad Album"}))
	writeFile(t, filepath.Join(albumDir, "badgenre.mp3"), fakeMp3([2]string{"TCON", "Bad Genre"}))

	files := ScanFiles(MediascanConf{
		MediaDirs:     []string{root},
		MediaExts:     []string{".mp3"},
		ExcludeTitle:  []string{"Live"},
		ExcludeArtist: []string{"Bad Artist"},
		ExcludeAlbum:  []string{"Bad Album"},
		ExcludeGenre:  []string{"Bad Genre"},
	})

	if len(files) != 1 {
		t.Fatalf("expected 1 media file, got %d", len(files))
	}
	if files[0].Path != keepPath {
		t.Errorf("got path %q, want %q", files[0].Path, keepPath)
	}
}

func TestGetMp3Duration_ReturnsZeroForMissingFile(t *testing.T) {
	if d := getMp3Duration(filepath.Join(t.TempDir(), "missing.mp3")); d != 0 {
		t.Errorf("got duration %v, want 0", d)
	}
}

func TestGetMp3Duration_ReturnsZeroForNonMp3Content(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fake.mp3")
	writeFile(t, path, []byte("not a complete mp3"))

	if d := getMp3Duration(path); d != 0 {
		t.Errorf("got duration %v, want 0", d)
	}
}
