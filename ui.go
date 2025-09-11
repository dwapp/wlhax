package main

import (
	"git.sr.ht/~rockorager/vaxis"
)

// UI represents the main UI interface
type UI interface {
	Draw(win vaxis.Window)
	HandleEvent(ev vaxis.Event) bool
}

// Tab represents a tab in the tab view
type Tab struct {
	Name    string
	Content UI
}

// TabView implements a tab-based interface
type TabView struct {
	tabs     []*Tab
	selected int
	win      vaxis.Window
}

// NewTabView creates a new tab view
func NewTabView() *TabView {
	return &TabView{
		tabs:     make([]*Tab, 0),
		selected: 0,
	}
}

// Add adds a new tab
func (tv *TabView) Add(content UI, name string) {
	tv.tabs = append(tv.tabs, &Tab{
		Name:    name,
		Content: content,
	})
}

// Selected returns the currently selected tab
func (tv *TabView) Selected() *Tab {
	if tv.selected >= 0 && tv.selected < len(tv.tabs) {
		return tv.tabs[tv.selected]
	}
	return nil
}

// NextTab moves to the next tab
func (tv *TabView) NextTab() {
	if len(tv.tabs) > 1 {
		tv.selected = (tv.selected + 1) % len(tv.tabs)
	}
}

// PrevTab moves to the previous tab
func (tv *TabView) PrevTab() {
	if len(tv.tabs) > 1 {
		tv.selected = (tv.selected - 1 + len(tv.tabs)) % len(tv.tabs)
	}
}

// Draw renders the tab view
func (tv *TabView) Draw(win vaxis.Window) {
	tv.win = win
	width, height := win.Size()
	
	if height < 2 {
		return
	}
	
	// Draw tab bar
	tv.drawTabBar(win.New(0, 0, width, 1))
	
	// Draw content
	if tab := tv.Selected(); tab != nil {
		contentWin := win.New(0, 1, width, height-1)
		tab.Content.Draw(contentWin)
	}
}

// drawTabBar draws the tab bar at the top
func (tv *TabView) drawTabBar(win vaxis.Window) {
	win.Clear()
	width, _ := win.Size()
	
	col := 0
	for i, tab := range tv.tabs {
		style := vaxis.Style{}
		if i == tv.selected {
			style.Attribute = vaxis.AttrReverse
		}
		
		tabName := " " + tab.Name + " "
		if col+len(tabName) > width {
			break
		}
		
		// Print each character with style
		for _, char := range []rune(tabName) {
			win.SetCell(col, 0, vaxis.Cell{
				Character: vaxis.Character{Grapheme: string(char), Width: 1},
				Style:     style,
			})
			col++
		}
	}
}

// HandleEvent handles events for the tab view
func (tv *TabView) HandleEvent(ev vaxis.Event) bool {
	if key, ok := ev.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyLeft), key.Matches('h'):
			tv.PrevTab()
			return true
		case key.Matches(vaxis.KeyRight), key.Matches('l'):
			tv.NextTab()
			return true
		}
	}
	
	// Pass event to current tab
	if tab := tv.Selected(); tab != nil {
		return tab.Content.HandleEvent(ev)
	}
	
	return false
}

// TextInput represents a simple text input widget
type TextInput struct {
	prompt string
	text   string
	cursor int
}

// NewTextInput creates a new text input
func NewTextInput(prompt string) *TextInput {
	return &TextInput{
		prompt: prompt,
		text:   "",
		cursor: 0,
	}
}

// SetText sets the text content
func (ti *TextInput) SetText(text string) {
	ti.text = text
	ti.cursor = len(text)
}

// Text returns the current text
func (ti *TextInput) Text() string {
	return ti.text
}

// Draw renders the text input
func (ti *TextInput) Draw(win vaxis.Window) {
	win.Clear()
	display := ti.prompt + ti.text
	
	// Print text using Segment
	win.Print(vaxis.Segment{Text: display})
	
	// Show cursor position
	cursorPos := len(ti.prompt) + ti.cursor
	width, _ := win.Size()
	if cursorPos < width {
		// We can show cursor by highlighting the character
		win.SetCell(cursorPos, 0, vaxis.Cell{
			Character: vaxis.Character{Grapheme: " "},
			Style:     vaxis.Style{Attribute: vaxis.AttrReverse},
		})
	}
}

// HandleEvent handles events for text input
func (ti *TextInput) HandleEvent(ev vaxis.Event) bool {
	if key, ok := ev.(vaxis.Key); ok {
		switch {
		case key.Matches(vaxis.KeyBackspace):
			if ti.cursor > 0 {
				ti.text = ti.text[:ti.cursor-1] + ti.text[ti.cursor:]
				ti.cursor--
			}
			return true
		case key.Matches(vaxis.KeyLeft):
			if ti.cursor > 0 {
				ti.cursor--
			}
			return true
		case key.Matches(vaxis.KeyRight):
			if ti.cursor < len(ti.text) {
				ti.cursor++
			}
			return true
		case key.Matches(vaxis.KeyHome):
			ti.cursor = 0
			return true
		case key.Matches(vaxis.KeyEnd):
			ti.cursor = len(ti.text)
			return true
		default:
			// Handle regular character input
			if len(key.Text) > 0 && key.Text != "" {
				ti.text = ti.text[:ti.cursor] + key.Text + ti.text[ti.cursor:]
				ti.cursor += len(key.Text)
				return true
			}
		}
	}
	return false
}
