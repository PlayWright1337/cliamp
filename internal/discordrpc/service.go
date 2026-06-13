package discordrpc

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hugolgst/rich-go/client"

	"cliamp/internal/playback"
)

const DefaultClientID = "1445120766562668688"

type Config struct {
	Enabled    bool
	ClientID   string
	LargeImage string
	SmallImage string
}

type Service struct {
	mu               sync.Mutex
	cfg              Config
	connected        bool
	lastKey          string
	lastState        playback.State
	nextLoginAttempt time.Time
}

func New(cfg Config) *Service {
	if !cfg.Enabled {
		return nil
	}
	if cfg.ClientID == "" {
		cfg.ClientID = DefaultClientID
	}
	return &Service{cfg: cfg}
}

func (s *Service) Update(state playback.State) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastState = state
	if state.Status == playback.StatusStopped {
		if s.connected {
			client.Logout()
			s.connected = false
		}
		s.lastKey = ""
		return
	}

	if !s.ensureConnectedLocked() {
		return
	}

	key := activityKey(state)
	if key == s.lastKey {
		return
	}
	if err := client.SetActivity(s.activity(state)); err != nil {
		s.connected = false
		s.nextLoginAttempt = time.Now().Add(30 * time.Second)
		return
	}
	s.lastKey = key
}

func (s *Service) Seeked(position time.Duration) {
	if s == nil {
		return
	}
	s.mu.Lock()
	state := s.lastState
	state.Position = position
	s.lastKey = ""
	s.mu.Unlock()
	s.Update(state)
}

func (s *Service) Close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connected {
		client.Logout()
		s.connected = false
	}
}

func (s *Service) ensureConnectedLocked() bool {
	if s.connected {
		return true
	}
	now := time.Now()
	if now.Before(s.nextLoginAttempt) {
		return false
	}
	if err := client.Login(s.cfg.ClientID); err != nil {
		s.nextLoginAttempt = now.Add(30 * time.Second)
		return false
	}
	s.connected = true
	return true
}

func (s *Service) activity(state playback.State) client.Activity {
	title := strings.TrimSpace(state.Track.Title)
	artist := strings.TrimSpace(state.Track.Artist)
	if title == "" {
		title = "Unknown Track"
	}
	status := "Playing"
	if state.Status == playback.StatusPaused {
		status = "Paused"
	}
	rpc := client.Activity{
		Details:    title,
		State:      artist,
		LargeImage: s.cfg.LargeImage,
		SmallImage: s.cfg.SmallImage,
		SmallText:  status,
	}
	if rpc.State == "" {
		rpc.State = status
	}
	if state.Status == playback.StatusPlaying {
		start := time.Now().Add(-state.Position)
		rpc.Timestamps = &client.Timestamps{Start: &start}
		if state.Track.Duration > 0 && state.Position < state.Track.Duration {
			end := start.Add(state.Track.Duration)
			rpc.Timestamps.End = &end
		}
	}
	if strings.HasPrefix(state.Track.URL, "http://") || strings.HasPrefix(state.Track.URL, "https://") {
		rpc.Buttons = []*client.Button{{Label: "Open Track", Url: state.Track.URL}}
	}
	return rpc
}

func activityKey(state playback.State) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%d\x00%s",
		state.Status,
		state.Track.Title,
		state.Track.Artist,
		state.Track.Album,
		state.Track.Duration,
		state.Track.URL,
	)
}
