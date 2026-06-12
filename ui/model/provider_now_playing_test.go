package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"cliamp/playlist"
	"cliamp/ui"
)

type providerNowPlayingFake struct{}

func (providerNowPlayingFake) Name() string { return "SoundCloud" }

func (providerNowPlayingFake) Playlists() ([]playlist.PlaylistInfo, error) {
	return []playlist.PlaylistInfo{
		{ID: "tracks", Name: "Tracks"},
		{ID: "likes", Name: "Likes"},
	}, nil
}

func (providerNowPlayingFake) Tracks(string) ([]playlist.Track, error) {
	return []playlist.Track{{Title: "Loaded", Path: "loaded.mp3"}}, nil
}

func TestDecorateProviderListsAddsNowPlayingForSoundCloud(t *testing.T) {
	prov := providerNowPlayingFake{}
	p := playlist.New()
	p.Replace([]playlist.Track{{Title: "Current", Path: "current.mp3"}})

	m := Model{
		provider: prov,
		playlist: p,
		providers: []ProviderEntry{
			{Key: "soundcloud", Name: "SoundCloud", Provider: prov},
		},
		provPillIdx: 0,
	}

	updated, _ := m.Update([]playlist.PlaylistInfo{
		{ID: "tracks", Name: "Tracks"},
		{ID: "likes", Name: "Likes"},
	})
	got := updated.(Model)

	if len(got.providerLists) != 3 {
		t.Fatalf("len(providerLists) = %d, want 3", len(got.providerLists))
	}
	if got.providerLists[0].ID != providerNowPlayingID {
		t.Fatalf("providerLists[0].ID = %q, want %q", got.providerLists[0].ID, providerNowPlayingID)
	}
	if got.providerLists[1].ID != "tracks" {
		t.Fatalf("providerLists[1].ID = %q, want tracks", got.providerLists[1].ID)
	}
}

func TestDecorateProviderListsAddsNowPlayingForSoundCloudByName(t *testing.T) {
	prov := providerNowPlayingFake{}
	p := playlist.New()
	p.Replace([]playlist.Track{{Title: "Current", Path: "current.mp3"}})

	m := Model{
		provider: prov,
		playlist: p,
	}

	got := m.decorateProviderLists([]playlist.PlaylistInfo{{ID: "tracks", Name: "Tracks"}})
	if len(got) != 2 || got[0].ID != providerNowPlayingID {
		t.Fatalf("decorated = %+v, want Now Playing prepended", got)
	}
}

func TestDecorateProviderListsSkipsNowPlayingWithoutSoundCloudPlaylist(t *testing.T) {
	prov := providerNowPlayingFake{}

	cases := []struct {
		name      string
		model     Model
		wantCount int
	}{
		{
			name: "other provider",
			model: Model{
				provider: prov,
				playlist: playlist.New(),
				providers: []ProviderEntry{
					{Key: "spotify", Name: "Spotify", Provider: prov},
				},
				provPillIdx: 0,
			},
			wantCount: 1,
		},
		{
			name: "empty playlist",
			model: Model{
				provider: prov,
				playlist: playlist.New(),
				providers: []ProviderEntry{
					{Key: "soundcloud", Name: "SoundCloud", Provider: prov},
				},
				provPillIdx: 0,
			},
			wantCount: 1,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.model.decorateProviderLists([]playlist.PlaylistInfo{{ID: "tracks", Name: "Tracks"}})
			if len(got) != tt.wantCount {
				t.Fatalf("len(decorated) = %d, want %d", len(got), tt.wantCount)
			}
			if got[0].ID == providerNowPlayingID {
				t.Fatalf("got Now Playing for %s", tt.name)
			}
		})
	}
}

func TestProviderNowPlayingSelectionReturnsToCurrentPlaylist(t *testing.T) {
	m := Model{
		focus:    focusProvider,
		playlist: playlist.New(),
		providerLists: []playlist.PlaylistInfo{
			{ID: providerNowPlayingID, Name: "Now Playing"},
			{ID: "tracks", Name: "Tracks"},
		},
	}

	cmd := m.handleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("cmd = %v, want nil", cmd)
	}
	if m.focus != focusPlaylist {
		t.Fatalf("focus = %v, want focusPlaylist", m.focus)
	}
	if m.provLoading {
		t.Fatal("provLoading = true, want false")
	}
	if m.activeProviderPlaylistID != "" {
		t.Fatalf("activeProviderPlaylistID = %q, want empty", m.activeProviderPlaylistID)
	}
}

func TestProviderPlaylistSelectionOpensTrackBrowser(t *testing.T) {
	prov := providerNowPlayingFake{}
	p := playlist.New()
	p.Replace([]playlist.Track{{Title: "Current", Path: "current.mp3"}})

	m := Model{
		focus:    focusProvider,
		provider: prov,
		playlist: p,
		providerLists: []playlist.PlaylistInfo{
			{ID: "tracks", Name: "Tracks"},
		},
	}

	cmd := m.handleKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("cmd = nil, want fetch command")
	}
	if !m.navBrowser.visible || !m.navBrowser.loading || m.navBrowser.screen != navBrowseScreenTracks {
		t.Fatalf("nav browser state = visible:%v loading:%v screen:%v, want loading track browser", m.navBrowser.visible, m.navBrowser.loading, m.navBrowser.screen)
	}
	if m.activeProviderPlaylistID != "" {
		t.Fatalf("activeProviderPlaylistID = %q, want empty", m.activeProviderPlaylistID)
	}
	if current, idx := m.playlist.Current(); current.Title != "Current" || idx != 0 {
		t.Fatalf("current = (%q,%d), want (\"Current\",0)", current.Title, idx)
	}

	msg := cmd()
	loaded, ok := msg.(tracksLoadedMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want tracksLoadedMsg", msg)
	}
	if loaded.name != "Tracks" || len(loaded.tracks) != 1 || loaded.tracks[0].Title != "Loaded" {
		t.Fatalf("loaded = %+v, want Tracks with Loaded", loaded)
	}
}

func TestTracksLoadedKeepsProviderTrackBrowserBackTarget(t *testing.T) {
	prov := providerNowPlayingFake{}
	p := playlist.New()
	p.Replace([]playlist.Track{{Title: "Current", Path: "current.mp3"}})

	m := Model{
		focus:    focusProvider,
		provider: prov,
		playlist: p,
		providers: []ProviderEntry{
			{Key: "soundcloud", Name: "SoundCloud", Provider: prov},
		},
		provPillIdx: 0,
		providerLists: []playlist.PlaylistInfo{
			{ID: "tracks", Name: "Tracks"},
		},
	}

	updated, _ := m.Update(tracksLoadedMsg{
		name:   "Tracks",
		tracks: []playlist.Track{{Title: "Loaded", Path: "loaded.mp3"}},
	})
	got := updated.(Model)
	if !got.navBrowser.providerTrack {
		t.Fatal("providerTrack = false, want true")
	}

	cmd := got.handleNavTrackListKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Fatalf("cmd = %v, want nil", cmd)
	}
	if got.navBrowser.visible {
		t.Fatal("navBrowser.visible = true, want false")
	}
	if got.focus != focusProvider {
		t.Fatalf("focus = %v, want focusProvider", got.focus)
	}
	if len(got.providerLists) != 2 || got.providerLists[0].ID != providerNowPlayingID {
		t.Fatalf("providerLists = %+v, want Now Playing prepended", got.providerLists)
	}
}

func TestNavBrowserCtrlKOpensKeymap(t *testing.T) {
	prov := providerNowPlayingFake{}
	player := &playbackFakeEngine{}
	m := Model{
		player:   player,
		playlist: playlist.New(),
		vis:      ui.NewVisualizer(float64(player.SampleRate())),
		navBrowser: navBrowserState{
			prov:          prov,
			visible:       true,
			providerTrack: true,
			mode:          navBrowseModeByAlbum,
			screen:        navBrowseScreenTracks,
			tracks:        []playlist.Track{{Title: "Loaded", Path: "loaded.mp3"}},
		},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	if cmd != nil {
		t.Fatalf("cmd = %v, want nil", cmd)
	}
	got := updated.(Model)
	if !got.keymap.visible {
		t.Fatal("keymap.visible = false, want true")
	}
}

func TestProviderTrackPlayAddsNowPlayingToSoundCloudMenu(t *testing.T) {
	prov := providerNowPlayingFake{}
	player := &playbackFakeEngine{}
	m := Model{
		player:   player,
		provider: prov,
		playlist: playlist.New(),
		vis:      ui.NewVisualizer(float64(player.SampleRate())),
		providers: []ProviderEntry{
			{Key: "soundcloud", Name: "SoundCloud", Provider: prov},
		},
		provPillIdx: 0,
		providerLists: []playlist.PlaylistInfo{
			{ID: "tracks", Name: "Tracks"},
			{ID: "likes", Name: "Likes"},
		},
		navBrowser: navBrowserState{
			prov:          prov,
			visible:       true,
			providerTrack: true,
			mode:          navBrowseModeByAlbum,
			screen:        navBrowseScreenTracks,
			tracks: []playlist.Track{
				{Title: "Loaded", Path: "https://example.com/loaded.mp3", Stream: true},
			},
		},
	}

	cmd := m.handleNavTrackListKey(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("cmd = nil, want playback command")
	}
	if len(m.providerLists) != 3 || m.providerLists[0].ID != providerNowPlayingID {
		t.Fatalf("providerLists = %+v, want Now Playing prepended", m.providerLists)
	}
	if current, idx := m.playlist.Current(); current.Title != "Loaded" || idx != 0 {
		t.Fatalf("current = (%q,%d), want (\"Loaded\",0)", current.Title, idx)
	}
}

func TestProviderTrackBrowserBackReturnsToProviderWithNowPlaying(t *testing.T) {
	prov := providerNowPlayingFake{}
	p := playlist.New()
	p.Replace([]playlist.Track{{Title: "Current", Path: "current.mp3"}})

	m := Model{
		focus:    focusProvider,
		provider: prov,
		playlist: p,
		providers: []ProviderEntry{
			{Key: "soundcloud", Name: "SoundCloud", Provider: prov},
		},
		provPillIdx: 0,
		providerLists: []playlist.PlaylistInfo{
			{ID: "tracks", Name: "Tracks"},
		},
		navBrowser: navBrowserState{
			prov:          prov,
			visible:       true,
			providerTrack: true,
			mode:          navBrowseModeByAlbum,
			screen:        navBrowseScreenTracks,
			tracks:        []playlist.Track{{Title: "Loaded", Path: "loaded.mp3"}},
		},
	}

	cmd := m.handleNavTrackListKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Fatalf("cmd = %v, want nil", cmd)
	}
	if m.navBrowser.visible {
		t.Fatal("navBrowser.visible = true, want false")
	}
	if m.focus != focusProvider {
		t.Fatalf("focus = %v, want focusProvider", m.focus)
	}
	if len(m.providerLists) != 2 || m.providerLists[0].ID != providerNowPlayingID {
		t.Fatalf("providerLists = %+v, want Now Playing prepended", m.providerLists)
	}
}
