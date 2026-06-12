package model

import (
	"reflect"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"cliamp/playlist"
	"cliamp/provider"
)

const providerNowPlayingID = "cliamp:now-playing"

// resetProviderNav resets provider navigation and search state to the top.
func (m *Model) resetProviderNav() {
	m.provCursor = 0
	m.provScroll = 0
	m.provLoading = true
	m.provSearch.active = false
	m.provSearch.query = ""
	m.provSearch.results = nil
	m.provSearch.cursor = 0
	m.provSearch.scroll = 0
}

func (m *Model) decorateProviderLists(lists []playlist.PlaylistInfo) []playlist.PlaylistInfo {
	lists = slices.DeleteFunc(slices.Clone(lists), func(info playlist.PlaylistInfo) bool {
		return info.ID == providerNowPlayingID
	})
	if !m.isSoundCloudProvider() || m.playlist == nil || m.playlist.Len() == 0 {
		return lists
	}
	out := make([]playlist.PlaylistInfo, 0, len(lists)+1)
	out = append(out, playlist.PlaylistInfo{ID: providerNowPlayingID, Name: "Now Playing", Section: "Cliamp"})
	out = append(out, lists...)
	return out
}

func (m *Model) isSoundCloudProvider() bool {
	if m.provider == nil {
		return false
	}
	if m.providerKey() == "soundcloud" {
		return true
	}
	return strings.EqualFold(m.provider.Name(), "SoundCloud")
}

func (m *Model) providerKey() string {
	if m.provPillIdx >= 0 && m.provPillIdx < len(m.providers) && sameProvider(m.providers[m.provPillIdx].Provider, m.provider) {
		return m.providers[m.provPillIdx].Key
	}
	for _, entry := range m.providers {
		if sameProvider(entry.Provider, m.provider) {
			return entry.Key
		}
	}
	return ""
}

func sameProvider(a, b playlist.Provider) bool {
	if a == nil || b == nil {
		return a == b
	}
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if av.Type() != bv.Type() || !av.Type().Comparable() {
		return false
	}
	return av.Interface() == bv.Interface()
}

// StartInProvider configures the model to begin in the provider browse view.
// Call this from main when no CLI tracks or pending URLs were given.
func (m *Model) StartInProvider() {
	if m.provider != nil {
		m.focus = focusProvider
		m.resetProviderNav()
	}
}

// switchProvider sets the active provider by pill index and fetches its playlists.
func (m *Model) switchProvider(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.providers) {
		return nil
	}
	m.provPillIdx = idx
	m.provider = m.providers[idx].Provider
	m.providerLists = nil
	m.provSignIn = false
	m.catalogBatch = catalogBatchState{}
	m.activeProviderPlaylistID = ""
	m.resetProviderNav()
	m.focus = focusProvider
	return fetchPlaylistsCmd(m.provider)
}

// quickSwitchProvider closes any browser overlays and jumps to the provider
// matched by key. Use the same Shift+letter shortcuts that switch providers
// from the main pane (S, N, P, J, E, Y, M, R, L). Returns nil when the key doesn't
// match a known provider.
func (m *Model) quickSwitchProvider(key string) tea.Cmd {
	provKey := providerKeyForShortcut(key)
	if provKey == "" {
		return nil
	}
	// Close any open overlays so the user lands on the provider pane.
	m.navBrowser.visible = false
	m.plManager.visible = false
	m.fileBrowser.visible = false
	return m.switchToProvider(provKey)
}

// providerKeyForShortcut maps the Shift+letter provider shortcuts to the
// config key used by switchToProvider, or "" when the key is unrelated.
func providerKeyForShortcut(key string) string {
	switch key {
	case "S":
		return "spotify"
	case "N":
		return "navidrome"
	case "P":
		return "plex"
	case "J":
		return "jellyfin"
	case "E":
		return "emby"
	case "Y":
		return "yt"
	case "M":
		return "netease"
	case "L":
		return "local"
	case "R":
		return "radio"
	}
	return ""
}

// switchToProvider finds a provider by config key and switches to it.
// Returns nil if the provider is not configured.
func (m *Model) switchToProvider(key string) tea.Cmd {
	for i, pe := range m.providers {
		if pe.Key == key {
			return m.switchProvider(i)
		}
	}
	return nil
}

// SetPendingURLs stores remote URLs (feeds, M3U) for async resolution after Init.
func (m *Model) SetPendingURLs(urls []string) {
	m.pendingURLs = urls
	m.feedLoading = len(urls) > 0
}

// findBrowseProvider returns the first provider that supports browsing
// (ArtistBrowser or AlbumBrowser), preferring the active provider.
func (m *Model) findBrowseProvider() playlist.Provider {
	return m.findProviderWith(func(p playlist.Provider) bool {
		if _, ok := p.(provider.ArtistBrowser); ok {
			return true
		}
		_, ok := p.(provider.AlbumBrowser)
		return ok
	})
}

func (m *Model) openNavBrowserWith(prov playlist.Provider) {
	m.navBrowser.prov = prov
	m.navBrowser.visible = true
	m.navBrowser.providerTrack = false
	m.navBrowser.mode = navBrowseModeMenu
	m.navBrowser.screen = navBrowseScreenList
	m.navBrowser.cursor = 0
	m.navBrowser.scroll = 0
	m.navBrowser.artists = nil
	m.navBrowser.albums = nil
	m.navBrowser.tracks = nil
	m.navBrowser.loading = false
	m.navBrowser.albumLoading = false
	m.navBrowser.albumDone = false
	m.navBrowser.searching = false
	m.navBrowser.search = ""
	m.navBrowser.searchIdx = nil
	m.navBrowser.selArtist = provider.ArtistInfo{}
	m.navBrowser.selAlbum = provider.AlbumInfo{}
	if ab, ok := prov.(provider.AlbumBrowser); ok {
		m.navBrowser.sortType = ab.DefaultAlbumSort()
	} else {
		m.navBrowser.sortType = ""
	}
}

func (m *Model) openProviderTrackBrowser(info playlist.PlaylistInfo) {
	m.navBrowser.prov = m.provider
	m.navBrowser.visible = true
	m.navBrowser.providerTrack = true
	m.navBrowser.mode = navBrowseModeByAlbum
	m.navBrowser.screen = navBrowseScreenTracks
	m.navBrowser.cursor = 0
	m.navBrowser.scroll = 0
	m.navBrowser.artists = nil
	m.navBrowser.albums = nil
	m.navBrowser.tracks = nil
	m.navBrowser.loading = true
	m.navBrowser.albumLoading = false
	m.navBrowser.albumDone = true
	m.navBrowser.searching = false
	m.navBrowser.search = ""
	m.navBrowser.searchIdx = nil
	m.navBrowser.selArtist = provider.ArtistInfo{}
	m.navBrowser.selAlbum = provider.AlbumInfo{Name: info.Name}
}

func (m *Model) refreshProviderNowPlaying() {
	if m.providerLists == nil {
		return
	}
	m.providerLists = m.decorateProviderLists(m.providerLists)
}

func (m *Model) ensureProviderListsDecorated() {
	if m.providerLists == nil {
		return
	}
	decorated := m.decorateProviderLists(m.providerLists)
	if len(decorated) != len(m.providerLists) {
		m.providerLists = decorated
		if m.provCursor >= len(m.providerLists) {
			m.provCursor = max(0, len(m.providerLists)-1)
		}
		if m.provSearch.active {
			m.updateProvSearch()
		}
	}
}

// navUpdateSearch rebuilds navSearchIdx from the current navSearch query
// against whichever list is active on the current nav screen.
func (m *Model) navUpdateSearch() {
	q := strings.ToLower(m.navBrowser.search)
	if q == "" {
		m.navBrowser.searchIdx = nil
		return
	}
	m.navBrowser.searchIdx = nil
	switch {
	case m.navBrowser.mode == navBrowseModeByArtist && m.navBrowser.screen == navBrowseScreenList,
		m.navBrowser.mode == navBrowseModeByArtistAlbum && m.navBrowser.screen == navBrowseScreenList:
		for i, a := range m.navBrowser.artists {
			if strings.Contains(strings.ToLower(a.Name), q) {
				m.navBrowser.searchIdx = append(m.navBrowser.searchIdx, i)
			}
		}
	case m.navBrowser.mode == navBrowseModeByAlbum && m.navBrowser.screen == navBrowseScreenList,
		m.navBrowser.mode == navBrowseModeByArtistAlbum && m.navBrowser.screen == navBrowseScreenAlbums:
		for i, a := range m.navBrowser.albums {
			if strings.Contains(strings.ToLower(a.Name), q) ||
				strings.Contains(strings.ToLower(a.Artist), q) {
				m.navBrowser.searchIdx = append(m.navBrowser.searchIdx, i)
			}
		}
	case m.navBrowser.screen == navBrowseScreenTracks:
		for i, t := range m.navBrowser.tracks {
			if strings.Contains(strings.ToLower(t.Title), q) ||
				strings.Contains(strings.ToLower(t.Artist), q) ||
				strings.Contains(strings.ToLower(t.Album), q) {
				m.navBrowser.searchIdx = append(m.navBrowser.searchIdx, i)
			}
		}
	}
}

// navClearSearch resets the nav search state.
func (m *Model) navClearSearch() {
	m.navBrowser.searching = false
	m.navBrowser.search = ""
	m.navBrowser.searchIdx = nil
	m.navBrowser.cursor = 0
	m.navBrowser.scroll = 0
}

// fetchNavArtistAllTracksCmd first fetches the artist's album list, then fetches
// all tracks across every album. This is used by the "By Artist" browse mode.
// The provider must implement both ArtistBrowser and AlbumTrackLoader.
func (m *Model) fetchNavArtistAllTracksCmd(ab provider.ArtistBrowser, artistID string) tea.Cmd {
	loader, _ := m.navBrowser.prov.(provider.AlbumTrackLoader)
	return func() tea.Msg {
		albums, err := ab.ArtistAlbums(artistID)
		if err != nil {
			return err
		}
		if loader == nil {
			return navTracksLoadedMsg(nil)
		}
		var all []playlist.Track
		for _, album := range albums {
			tracks, err := loader.AlbumTracks(album.ID)
			if err != nil {
				return err
			}
			all = append(all, tracks...)
		}
		return navTracksLoadedMsg(all)
	}
}
