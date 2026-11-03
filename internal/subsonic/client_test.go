package subsonic

import (
	"context"
	"net/http"
	"testing"

	"github.com/rairulyle/bitchord-selfhosted-addon/internal/fakes"
	"github.com/rairulyle/bitchord-selfhosted-addon/internal/media"
)

func newClient(fake *fakes.Navidrome) *Client {
	return New(Options{BaseURL: fake.URL, Username: fakes.NavidromeUser, Password: fakes.NavidromePassword})
}

func TestVersion(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	version, err := newClient(fake).Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if version != "1.16.1" {
		t.Fatalf("version = %q, want 1.16.1", version)
	}
}

func TestWrongPasswordIsUnauthorized(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	_, err := New(Options{BaseURL: fake.URL, Username: fakes.NavidromeUser, Password: "wrong-password"}).Version(context.Background())
	if err == nil || !isUnauthorized(err) {
		t.Fatalf("err = %v, want unauthorized", err)
	}
}

func TestLibraries(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	libs, err := newClient(fake).Libraries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 1 || libs[0].ID != fakes.NavidromeMusicFolder || libs[0].Name != fakes.NavidromeMusicFolderName {
		t.Fatalf("libraries = %+v", libs)
	}
}

func TestAllTracksSearch3(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	tracks, err := newClient(fake).AllTracks(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3: %+v", len(tracks), tracks)
	}
	byID := map[string]media.Track{}
	for _, tr := range tracks {
		byID[tr.ID] = tr
	}
	flac := byID["s-1"]
	if flac.Title != "New Religion" || flac.Codec != "flac" || !flac.Playable() {
		t.Fatalf("flac track = %+v", flac)
	}
	if !media.Lossless(flac.Codec) {
		t.Fatal("flac should be lossless")
	}
	if flac.SampleRate != 48000 || flac.BitDepth != 24 {
		t.Fatalf("hi-res metadata missing: %+v", flac)
	}
	if flac.AlbumArtist != "All Time Low" {
		t.Fatalf("album artist = %q, want All Time Low", flac.AlbumArtist)
	}
	mp3 := byID["s-2"]
	if mp3.Codec != "mp3" || mp3.ArtRef != "" {
		t.Fatalf("mp3 track = %+v", mp3)
	}
	m4a := byID["s-3"]
	if m4a.Codec != "aac" || media.Lossless(m4a.Codec) {
		t.Fatalf("m4a should map to aac (not lossless): %+v", m4a)
	}
}

func TestAllTracksFallbackViaAlbums(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	fake.DisableSearch3 = true // Force fallback to album walk
	tracks, err := newClient(fake).AllTracks(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 3 {
		t.Fatalf("got %d tracks via album walk, want 3: %+v", len(tracks), tracks)
	}
	byID := map[string]media.Track{}
	for _, tr := range tracks {
		byID[tr.ID] = tr
	}
	if _, ok := byID["s-1"]; !ok {
		t.Errorf("missing s-1 from album walk")
	}
	if _, ok := byID["s-2"]; !ok {
		t.Errorf("missing s-2 from album walk")
	}
	if _, ok := byID["s-3"]; !ok {
		t.Errorf("missing s-3 from album walk")
	}
}

func TestTrack(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	tr, err := newClient(fake).Track(context.Background(), "s-1")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Title != "New Religion" || tr.DurationSec != 184 {
		t.Fatalf("track = %+v", tr)
	}
	if _, err := newClient(fake).Track(context.Background(), "missing"); err != media.ErrNotFound {
		t.Fatalf("missing track err = %v, want ErrNotFound", err)
	}
}

func TestOpenFileAndArt(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	c := newClient(fake)
	tracks, err := c.AllTracks(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	var withArt media.Track
	for _, tr := range tracks {
		if tr.ArtRef != "" {
			withArt = tr
		}
	}
	file, err := c.OpenFile(context.Background(), withArt, http.MethodGet, nil)
	if err != nil {
		t.Fatal(err)
	}
	file.Body.Close()
	if file.StatusCode != http.StatusOK {
		t.Fatalf("file status = %d", file.StatusCode)
	}
	art, err := c.OpenArt(context.Background(), withArt)
	if err != nil {
		t.Fatal(err)
	}
	art.Body.Close()
	if art.StatusCode != http.StatusOK {
		t.Fatalf("art status = %d", art.StatusCode)
	}
}

func TestSplitJoinCredential(t *testing.T) {
	user, pass := SplitCredential("lyle:pa:ss")
	if user != "lyle" || pass != "pa:ss" {
		t.Fatalf("SplitCredential = %q, %q", user, pass)
	}
	if JoinCredential("lyle", "pa:ss") != "lyle:pa:ss" {
		t.Fatal("JoinCredential round trip failed")
	}
}

func TestResolveFolderByName(t *testing.T) {
	fake := fakes.NewNavidrome(t)
	if got := newClient(fake).resolveFolderID(context.Background(), "Music"); got != fakes.NavidromeMusicFolder {
		t.Fatalf("resolveFolderID(Music) = %q, want %s", got, fakes.NavidromeMusicFolder)
	}
	if got := newClient(fake).resolveFolderID(context.Background(), ""); got != "" {
		t.Fatalf("resolveFolderID(empty) = %q, want empty", got)
	}
}
