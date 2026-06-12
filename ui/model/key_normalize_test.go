package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNormalizeHotkeyMsgRussianLayout(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
		want string
	}{
		{"lower", tea.KeyPressMsg{Text: "о", Code: 'о'}, "j"},
		{"upper", tea.KeyPressMsg{Text: "О", Code: 'о', Mod: tea.ModShift}, "J"},
		{"ctrl", tea.KeyPressMsg{Code: 'а', Mod: tea.ModCtrl}, "ctrl+f"},
		{"base code", tea.KeyPressMsg{Text: "ф", Code: 'ф', BaseCode: 'a'}, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeHotkeyMsg(tt.msg).String(); got != tt.want {
				t.Fatalf("normalizeHotkeyMsg().String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeyTextInputActiveKeepsSearchTextLiteral(t *testing.T) {
	m := Model{search: searchState{active: true}}
	if !m.keyTextInputActive() {
		t.Fatal("keyTextInputActive() = false, want true")
	}
}
