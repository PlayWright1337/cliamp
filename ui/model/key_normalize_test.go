package model

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNormalizeHotkeyMsgLayoutIndependentPhysicalCode(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
		want string
	}{
		{"base code", tea.KeyPressMsg{Text: "ф", Code: 'ф', BaseCode: 'a'}, "a"},
		{"shifted code", tea.KeyPressMsg{Text: "あ", Code: 'あ', ShiftedCode: 'j'}, "j"},
		{"ctrl base code", tea.KeyPressMsg{Code: 'а', BaseCode: 'f', Mod: tea.ModCtrl}, "ctrl+f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeHotkeyMsg(tt.msg).String(); got != tt.want {
				t.Fatalf("normalizeHotkeyMsg().String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeHotkeyMsgCyrillicFallback(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
		want string
	}{
		{"lower", tea.KeyPressMsg{Text: "о", Code: 'о'}, "j"},
		{"upper", tea.KeyPressMsg{Text: "О", Code: 'о', Mod: tea.ModShift}, "J"},
		{"ctrl", tea.KeyPressMsg{Code: 'а', Mod: tea.ModCtrl}, "ctrl+f"},
		{"ukrainian non-ambiguous", tea.KeyPressMsg{Text: "ї", Code: 'ї'}, "]"},
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
