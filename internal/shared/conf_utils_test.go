package shared

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestResolveScanConfig(t *testing.T) {
	root := t.TempDir()
	mediaDir := filepath.Join(root, "data", "Music")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		t.Fatalf("mkdir music: %v", err)
	}

	confPath := filepath.Join(root, "mediascan-config.yml")
	content := `mediaDirs:
  - data/Music
mediaExts:
  - .mp3
`
	if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf, err := ResolveScanConfig(confPath, root)
	if err != nil {
		t.Fatalf("ResolveScanConfig returned error: %v", err)
	}
	if len(conf.MediaDirs) != 1 || conf.MediaDirs[0] != mediaDir {
		t.Fatalf("unexpected resolved media dirs: %#v", conf.MediaDirs)
	}
	if len(conf.MediaExts) != 1 || conf.MediaExts[0] != ".mp3" {
		t.Fatalf("unexpected media exts: %#v", conf.MediaExts)
	}
}

func TestParseCommandArgs_OverwriteFlags(t *testing.T) {
	mode, args, err := ParseCommandArgs([]string{"prog", "config.yml", "out.yml", "--overwrite-existing"}, 2, 3)
	if err != nil {
		t.Fatalf("ParseCommandArgs returned error: %v", err)
	}
	if mode != OverwriteExisting {
		t.Fatalf("mode = %v, want %v", mode, OverwriteExisting)
	}
	if len(args) != 2 || args[0] != "config.yml" || args[1] != "out.yml" {
		t.Fatalf("unexpected positional args: %#v", args)
	}

	mode, _, err = ParseCommandArgs([]string{"prog", "--overwrite-newer", "config.yml", "out.yml"}, 2, 3)
	if err != nil {
		t.Fatalf("ParseCommandArgs returned error: %v", err)
	}
	if mode != OverwriteNewer {
		t.Fatalf("mode = %v, want %v", mode, OverwriteNewer)
	}

	_, _, err = ParseCommandArgs([]string{"prog", "--overwrite-existing", "--overwrite-newer", "config.yml"}, 1, 2)
	if err == nil {
		t.Fatal("expected mutual exclusion error")
	}
}

func TestShouldOverwriteFile(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.txt")
	destPath := filepath.Join(dir, "dest.txt")

	if err := os.WriteFile(sourcePath, []byte("new"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(destPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write dest: %v", err)
	}

	if err := os.Chtimes(destPath, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("set dest time: %v", err)
	}
	if err := os.Chtimes(sourcePath, time.Now(), time.Now()); err != nil {
		t.Fatalf("set source time: %v", err)
	}

	should, err := ShouldOverwriteFile(destPath, sourcePath, OverwriteExisting)
	if err != nil {
		t.Fatalf("ShouldOverwriteFile returned error: %v", err)
	}
	if !should {
		t.Fatal("expected overwrite for older destination when overwrite-existing is set")
	}

	if err := os.Chtimes(destPath, time.Now().Add(time.Hour), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("set newer dest time: %v", err)
	}
	should, err = ShouldOverwriteFile(destPath, sourcePath, OverwriteExisting)
	if err != nil {
		t.Fatalf("ShouldOverwriteFile returned error: %v", err)
	}
	if should {
		t.Fatal("expected skip for newer destination when overwrite-existing is set")
	}

	should, err = ShouldOverwriteFile(destPath, sourcePath, OverwriteNewer)
	if err != nil {
		t.Fatalf("ShouldOverwriteFile returned error: %v", err)
	}
	if !should {
		t.Fatal("expected overwrite for newer destination when overwrite-newer is set")
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
