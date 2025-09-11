package ui

import (
	"sync"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/mattn/go-runewidth"
)

type Tabs struct {
	tabs       []*Tab
	TabStrip   *TabStrip
	TabContent *TabContent
	curIndex   int
	history    []int
	m          sync.Mutex
	CloseTab   func(index int)
}

type Tab struct {
	Content Drawable
	Name    string
	title   string
}

func (t *Tab) SetTitle(s string) {
	t.title = s
}

func (t *Tab) displayName() string {
	name := t.Name
	if t.title != "" {
		name = t.title
	}
	return name
}

type (
	TabStrip   Tabs
	TabContent Tabs
)

func NewTabs() *Tabs {
	tabs := &Tabs{}
	tabs.TabStrip = (*TabStrip)(tabs)
	tabs.TabContent = (*TabContent)(tabs)
	tabs.history = []int{}
	return tabs
}

func (tabs *Tabs) Add(content Drawable, name string, background bool) *Tab {
	tab := &Tab{
		Content: content,
		Name:    name,
	}
	tabs.tabs = append(tabs.tabs, tab)
	if !background {
		tabs.selectPriv(len(tabs.tabs) - 1)
	}
	return tab
}

func (tabs *Tabs) Remove(content Drawable) {
	tabs.m.Lock()
	defer tabs.m.Unlock()
	indexToRemove := -1
	for i, tab := range tabs.tabs {
		if tab.Content == content {
			tabs.tabs = append(tabs.tabs[:i], tabs.tabs[i+1:]...)
			tabs.removeHistory(i)
			indexToRemove = i
			break
		}
	}
	if indexToRemove < 0 {
		return
	}

	if tabs.curIndex == indexToRemove {
		index := 0
		if len(tabs.history) > 0 {
			index = tabs.history[0]
		}
		tabs.selectPriv(index)
	} else if tabs.curIndex > indexToRemove {
		tabs.curIndex--
	}
}

func (tabs *Tabs) Selected() *Tab {
	if tabs.curIndex >= 0 && tabs.curIndex < len(tabs.tabs) {
		return tabs.tabs[tabs.curIndex]
	}
	return nil
}

func (tabs *Tabs) Select(index int) {
	tabs.selectPriv(index)
}

func (tabs *Tabs) selectPriv(index int) {
	if index < 0 || index >= len(tabs.tabs) {
		return
	}
	
	if tabs.curIndex != index {
		tabs.pushHistory(tabs.curIndex)
		tabs.curIndex = index
		tabs.removeHistory(index)
	}
}

func (tabs *Tabs) NextTab() {
	next := tabs.curIndex + 1
	if next >= len(tabs.tabs) {
		next = 0
	}
	tabs.Select(next)
}

func (tabs *Tabs) PrevTab() {
	next := tabs.curIndex - 1
	if next < 0 {
		next = len(tabs.tabs) - 1
	}
	tabs.Select(next)
}

func (tabs *Tabs) pushHistory(index int) {
	if index < 0 || index >= len(tabs.tabs) {
		return
	}
	
	for i, item := range tabs.history {
		if item == index {
			tabs.history = append(tabs.history[:i], tabs.history[i+1:]...)
			break
		}
	}
	tabs.history = append([]int{index}, tabs.history...)
}

func (tabs *Tabs) removeHistory(index int) {
	for i, item := range tabs.history {
		if item == index {
			tabs.history = append(tabs.history[:i], tabs.history[i+1:]...)
		} else if item > index {
			tabs.history[i]--
		}
	}
}

// TabStrip methods
func (strip *TabStrip) Draw(ctx *Context) {
	tabs := (*Tabs)(strip)
	x := 0
	for i, tab := range tabs.tabs {
		name := tab.displayName()
		width := runewidth.StringWidth(name) + 2 // padding
		
		style := vaxis.Style{}
		if i == tabs.curIndex {
			style.Attribute = vaxis.AttrReverse
		}
		
		if x+width <= ctx.Width() {
			ctx.Printf(x, 0, style, " %s ", name)
			x += width
		}
	}
	
	// Fill remaining space
	if x < ctx.Width() {
		ctx.Fill(x, 0, ctx.Width()-x, 1, ' ', vaxis.Style{})
	}
}

func (strip *TabStrip) Invalidate() {
	Invalidate()
}

func (strip *TabStrip) Event(event vaxis.Event) bool {
	return false
}

func (strip *TabStrip) Focus(focus bool) {
	// Tab strip doesn't receive focus
}

// TabContent methods
func (content *TabContent) Draw(ctx *Context) {
	tabs := (*Tabs)(content)
	if tab := tabs.Selected(); tab != nil && tab.Content != nil {
		tab.Content.Draw(ctx)
	}
}

func (content *TabContent) Invalidate() {
	Invalidate()
}

func (content *TabContent) Event(event vaxis.Event) bool {
	tabs := (*Tabs)(content)
	if tab := tabs.Selected(); tab != nil {
		if interactive, ok := tab.Content.(Interactive); ok {
			return interactive.Event(event)
		}
	}
	return false
}

func (content *TabContent) Focus(focus bool) {
	tabs := (*Tabs)(content)
	if tab := tabs.Selected(); tab != nil {
		if interactive, ok := tab.Content.(Interactive); ok {
			interactive.Focus(focus)
		}
	}
}
