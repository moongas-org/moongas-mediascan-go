package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMediaDirs_WithRootDir(t *testing.T) {
	root := t.TempDir()
	musicDir := filepath.Join(root, "data", "Music")
	otherDir := filepath.Join(root, "data", "MusicOther")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		t.Fatalf("mkdir music: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatalf("mkdir other: %v", err)
	}

	resolved, err := ResolveMediaDirs(root, []string{"data/Music", "data/MusicOther"})
	if err != nil {
		t.Fatalf("ResolveMediaDirs returned error: %v", err)
	}

	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved dirs, got %d", len(resolved))
	}
	if resolved[0] != musicDir || resolved[1] != otherDir {
		t.Fatalf("unexpected resolved dirs: %#v", resolved)
	}
}

func TestResolveMediaDirs_RejectsMissingRootDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := ResolveMediaDirs(missing, []string{"data/Music"})
	if err == nil {
		t.Fatal("expected error for missing root dir")
	}
}

func TestResolveMediaDirs_RejectsMissingJoinedDir(t *testing.T) {
	root := t.TempDir()

	_, err := ResolveMediaDirs(root, []string{"data/Music"})
	if err == nil {
		t.Fatal("expected error for missing assembled media dir")
	}
}

func TestLoadConf(t *testing.T) {
	dir := t.TempDir()
	confPath := filepath.Join(dir, "mediascan-config.yml")
	content := `mediaDirs:
  - /data/Music
  - /data/MusicOther
mediaExts:
  - .mp3
  - .m4a
excludePaths:
  - ExcludeThisArtistDir
excludeTitle:
  - Live
excludeArtist:
  - Some Artist
excludeAlbumArtist: []
excludeAlbum:
  - Some Album
excludeGenre: []
sortBy: year
getMp3Duration: true
`
	if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf := LoadConf(confPath)

	if len(conf.MediaDirs) != 2 || conf.MediaDirs[0] != "/data/Music" || conf.MediaDirs[1] != "/data/MusicOther" {
		t.Errorf("unexpected media dirs: %#v", conf.MediaDirs)
	}
	if len(conf.MediaExts) != 2 || conf.MediaExts[0] != ".mp3" || conf.MediaExts[1] != ".m4a" {
		t.Errorf("unexpected media exts: %#v", conf.MediaExts)
	}
	if len(conf.ExcludePaths) != 1 || conf.ExcludePaths[0] != "ExcludeThisArtistDir" {
		t.Errorf("unexpected exclude paths: %#v", conf.ExcludePaths)
	}
	if len(conf.ExcludeTitle) != 1 || conf.ExcludeTitle[0] != "Live" {
		t.Errorf("unexpected exclude title: %#v", conf.ExcludeTitle)
	}
	if len(conf.ExcludeArtist) != 1 || conf.ExcludeArtist[0] != "Some Artist" {
		t.Errorf("unexpected exclude artist: %#v", conf.ExcludeArtist)
	}
	if len(conf.ExcludeAlbum) != 1 || conf.ExcludeAlbum[0] != "Some Album" {
		t.Errorf("unexpected exclude album: %#v", conf.ExcludeAlbum)
	}
	if conf.SortBy != "year" {
		t.Errorf("got sortBy %q, want %q", conf.SortBy, "year")
	}
	if !conf.GetMp3Duration {
		t.Error("expected getMp3Duration to be true")
	}
}
