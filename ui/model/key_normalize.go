package model

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

func (m *Model) keyTextInputActive() bool {
	return m.jumping ||
		m.urlInputting ||
		m.search.active ||
		m.provSearch.active ||
		(m.netSearch.active && m.netSearch.screen == netSearchInput) ||
		(m.spotSearch.visible && (m.spotSearch.screen == spotSearchInput || m.spotSearch.screen == spotSearchNewName)) ||
		(m.navBrowser.visible && m.navBrowser.searching) ||
		(m.fileBrowser.visible && m.fileBrowser.searching) ||
		(m.keymap.visible && m.keymap.searching) ||
		(m.plManager.visible && (m.plManager.filtering || m.plManager.screen == plMgrScreenNewName || m.plManager.screen == plMgrScreenRename))
}

func normalizeHotkeyMsg(msg tea.KeyPressMsg) tea.KeyPressMsg {
	key := msg.Key()
	if key.BaseCode != 0 && key.BaseCode < 128 && unicode.IsPrint(key.BaseCode) {
		key = keyWithLatinCode(key, key.BaseCode)
		return tea.KeyPressMsg(key)
	}
	if len(key.Text) == 1 {
		if mapped, ok := qwertyRune([]rune(key.Text)[0]); ok {
			key = keyWithLatinCode(key, mapped)
			return tea.KeyPressMsg(key)
		}
	}
	if mapped, ok := qwertyRune(key.Code); ok {
		key = keyWithLatinCode(key, mapped)
		return tea.KeyPressMsg(key)
	}
	return msg
}

func keyWithLatinCode(key tea.Key, code rune) tea.Key {
	shifted := unicode.IsUpper(code) || key.Mod.Contains(tea.ModShift)
	code = unicode.ToLower(code)
	key.Code = code
	key.BaseCode = code
	key.ShiftedCode = unicode.ToUpper(code)
	if key.Mod.Contains(tea.ModCtrl) || key.Mod.Contains(tea.ModAlt) || key.Mod.Contains(tea.ModMeta) {
		key.Text = ""
		return key
	}
	text := string(code)
	if shifted {
		text = strings.ToUpper(text)
	}
	key.Text = text
	return key
}

func qwertyRune(r rune) (rune, bool) {
	upper := unicode.IsUpper(r)
	r = unicode.ToLower(r)
	mapped, ok := russianQwerty[r]
	if !ok {
		return 0, false
	}
	if upper && mapped >= 'a' && mapped <= 'z' {
		mapped = unicode.ToUpper(mapped)
	}
	return mapped, true
}

var russianQwerty = map[rune]rune{
	'й': 'q',
	'ц': 'w',
	'у': 'e',
	'к': 'r',
	'е': 't',
	'н': 'y',
	'г': 'u',
	'ш': 'i',
	'щ': 'o',
	'з': 'p',
	'х': '[',
	'ъ': ']',
	'ф': 'a',
	'ы': 's',
	'в': 'd',
	'а': 'f',
	'п': 'g',
	'р': 'h',
	'о': 'j',
	'л': 'k',
	'д': 'l',
	'ж': ';',
	'э': '\'',
	'я': 'z',
	'ч': 'x',
	'с': 'c',
	'м': 'v',
	'и': 'b',
	'т': 'n',
	'ь': 'm',
	'б': ',',
	'ю': '.',
	'ё': '`',
}
