package main

import (
	"unicode/utf8"

	gsq "github.com/kballard/go-shellquote"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var arrowLeft = '←'
var arrowRight = '→'

const preferred_horizontal_threshold = 5

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func voffset(text []byte, boffset int) int {
	text = text[:boffset]
	res := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]
		res += runewidth.RuneWidth(r)
	}
	return res
}

type CommandLine struct {
	text           []byte
	line_voffset   int
	cursor_boffset int // cursor offset in bytes
	cursor_voffset int // visual cursor offset in termbox cells
}

// Draws the CommandLine in the given location, 'h' is not used at the moment
func (eb *CommandLine) Draw(x, y, w, h int) {
	eb.AdjustVOffset(w)

	const coldef = termbox.ColorDefault
	const colred = termbox.ColorRed

	fill(x, y, w, h, termbox.Cell{Ch: ' '})

	t := eb.text
	lx := 0
	for {
		rx := lx - eb.line_voffset
		if len(t) == 0 {
			break
		}

		if rx >= w {
			termbox.SetCell(x+w-1, y, arrowRight,
				colred, coldef)
			break
		}

		r, size := utf8.DecodeRune(t)
		if rx >= 0 {
			termbox.SetCell(x+rx, y, r, coldef, coldef)
		}
		lx += runewidth.RuneWidth(r)
		t = t[size:]
	}

	if eb.line_voffset != 0 {
		termbox.SetCell(x, y, arrowLeft, colred, coldef)
	}
}

// Adjusts line visual offset to a proper value depending on width
func (eb *CommandLine) AdjustVOffset(width int) {
	ht := preferred_horizontal_threshold
	max_h_threshold := (width - 1) / 2
	if ht > max_h_threshold {
		ht = max_h_threshold
	}

	threshold := width - 1
	if eb.line_voffset != 0 {
		threshold = width - ht
	}
	if eb.cursor_voffset-eb.line_voffset >= threshold {
		eb.line_voffset = eb.cursor_voffset + (ht - width + 1)
	}

	if eb.line_voffset != 0 && eb.cursor_voffset-eb.line_voffset < ht {
		eb.line_voffset = eb.cursor_voffset - ht
		if eb.line_voffset < 0 {
			eb.line_voffset = 0
		}
	}
}

func (eb *CommandLine) MoveCursorTo(boffset int) {
	eb.cursor_boffset = boffset
	eb.cursor_voffset = voffset(eb.text, boffset)
}

func (eb *CommandLine) RuneUnderCursor() (rune, int) {
	return utf8.DecodeRune(eb.text[eb.cursor_boffset:])
}

func (eb *CommandLine) RuneBeforeCursor() (rune, int) {
	return utf8.DecodeLastRune(eb.text[:eb.cursor_boffset])
}

func (eb *CommandLine) MoveCursorOneRuneBackward() {
	if eb.cursor_boffset == 0 {
		return
	}
	_, size := eb.RuneBeforeCursor()
	eb.MoveCursorTo(eb.cursor_boffset - size)
}

func (eb *CommandLine) MoveCursorOneRuneForward() {
	if eb.cursor_boffset == len(eb.text) {
		return
	}
	_, size := eb.RuneUnderCursor()
	eb.MoveCursorTo(eb.cursor_boffset + size)
}

func (eb *CommandLine) DeleteRuneBackward() {
	if eb.cursor_boffset == 0 {
		return
	}

	eb.MoveCursorOneRuneBackward()
	_, size := eb.RuneUnderCursor()
	eb.text = byteSliceRemove(eb.text, eb.cursor_boffset, eb.cursor_boffset+size)
}

func (eb *CommandLine) DeleteAll() {
	eb.MoveCursorTo(0)
	eb.text = byteSliceRemove(eb.text, 0, len(eb.text))
}

func (eb *CommandLine) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	eb.text = byteSliceInsert(eb.text, eb.cursor_boffset, buf[:n])
	eb.MoveCursorOneRuneForward()
}

func (eb *CommandLine) Redraw() {
	w, h := termbox.Size()
	termbox.SetCell(0, h-1, ':', termbox.ColorDefault, termbox.ColorDefault)
	eb.Draw(1, h-1, w-1, 1)
	termbox.SetCursor(1+eb.CursorX(), h-1)

}

func (cl *CommandLine) Run() <-chan []string {
	ch := make(chan []string)
	go func() {
	mainloop:
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					close(ch)
					break mainloop
				case termbox.KeyArrowLeft, termbox.KeyCtrlB:
					cl.MoveCursorOneRuneBackward()
				case termbox.KeyArrowRight, termbox.KeyCtrlF:
					cl.MoveCursorOneRuneForward()
				case termbox.KeyBackspace, termbox.KeyBackspace2:
					cl.DeleteRuneBackward()
				case termbox.KeyTab:
				case termbox.KeySpace:
					cl.InsertRune(' ')
				case termbox.KeyHome, termbox.KeyCtrlA:
					cl.MoveCursorTo(0)
				case termbox.KeyEnd, termbox.KeyCtrlE:
					cl.MoveCursorTo(len(cl.text))
				case termbox.KeyEnter:
					tokens, err := gsq.Split(string(cl.text))
					if err != nil {
						panic(err)
					}
					if len(tokens) == 0 {
						break
					}
					if tokens[0] == "exit" {
						close(ch)
						break mainloop
					}
					ch <- tokens
					cl.DeleteAll()
				default:
					if ev.Ch != 0 {
						cl.InsertRune(ev.Ch)
					}
				}
			case termbox.EventError:
				panic(ev.Err)
			}
			cl.Redraw()
			termbox.Flush()
		}
	}()
	return ch
}

// Please, keep in mind that cursor depends on the value of line_voffset, which
// is being set on Draw() call, so.. call this method after Draw() one.
func (eb *CommandLine) CursorX() int {
	return eb.cursor_voffset - eb.line_voffset
}
