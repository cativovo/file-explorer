package main

import (
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		window := new(app.Window)
		appHeight := unit.Dp(600)
		appWidth := unit.Dp(600)
		window.Option(
			app.MinSize(appWidth, appHeight),
			app.MaxSize(appWidth, appHeight),
		)

		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {
	var ops op.Ops
	ma := newMyApp()

	go ma.update(window)

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			ma.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

type msg any

type changeDirMsg struct {
	path []string
}

type item struct {
	btn   widget.Clickable
	entry os.DirEntry
}

type myApp struct {
	items   []item
	list    widget.List
	theme   *material.Theme
	path    []string
	msgChan chan msg
}

func newMyApp() myApp {
	ma := myApp{
		theme:   material.NewTheme(),
		msgChan: make(chan msg, 1),
	}
	ma.list.Axis = layout.Vertical
	// TODO: use default $HOME path
	p := []string{"/", "home"}
	ma.changeDir(p)
	return ma
}

func (m *myApp) changeDir(p []string) {
	entries, err := os.ReadDir(filepath.Join(p...))
	if err != nil {
		panic(err)
	}
	m.path = p

	items := make([]item, 0, len(entries))
	for _, v := range entries {
		items = append(items, item{entry: v})
	}
	m.items = items
}

func (m *myApp) update(window *app.Window) {
	for v := range m.msgChan {
		switch v := v.(type) {
		case changeDirMsg:
			m.changeDir(v.path)
		}
		window.Invalidate()
	}
}

func (m *myApp) layout(gtx layout.Context) {
	layout.Flex{Axis: layout.Horizontal}.Layout(
		gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(m.theme, &m.list).Layout(gtx, len(m.items), func(gtx layout.Context, index int) layout.Dimensions {
				item := &m.items[index]
				if item.btn.Clicked(gtx) {
					if item.entry.IsDir() {
						m.msgChan <- changeDirMsg{path: append(m.path, item.entry.Name())}
					}
				}
				return material.Clickable(gtx, &item.btn, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(12)).Layout(gtx, material.H6(m.theme, item.entry.Name()).Layout)
				})
			})
		}),
	)
}
