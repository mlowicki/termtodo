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

type CommandLine struct {
	text             []byte
	lineCellOffset   int
	cursorByteOffset int
	cursorCellOffset int
}

// Draws the CommandLine in the given location, 'h' is not used at the moment
func (eb *CommandLine) Draw(x, y, w, h int) {
	eb.AdjustLineCellOffset(w)

	const coldef = termbox.ColorDefault
	const colred = termbox.ColorRed

	fill(x, y, w, h, termbox.Cell{Ch: ' '})

	t := eb.text
	lx := 0
	for {
		rx := lx - eb.lineCellOffset
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

	if eb.lineCellOffset != 0 {
		termbox.SetCell(x, y, arrowLeft, colred, coldef)
	}
}

// AdjustLineCellOffset sets line visual offset based on desired line width.
func (eb *CommandLine) AdjustLineCellOffset(width int) {
	ht := preferred_horizontal_threshold
	max_h_threshold := (width - 1) / 2
	if ht > max_h_threshold {
		ht = max_h_threshold
	}

	threshold := width - 1
	if eb.lineCellOffset != 0 {
		threshold = width - ht
	}
	if eb.cursorCellOffset-eb.lineCellOffset >= threshold {
		eb.lineCellOffset = eb.cursorCellOffset + (ht - width + 1)
	}

	if eb.lineCellOffset != 0 && eb.cursorCellOffset-eb.lineCellOffset < ht {
		eb.lineCellOffset = eb.cursorCellOffset - ht
		if eb.lineCellOffset < 0 {
			eb.lineCellOffset = 0
		}
	}
}

func (eb *CommandLine) MoveCursorTo(boffset int) {
	eb.cursorByteOffset = boffset
	eb.cursorCellOffset = wcwidth(eb.text[:boffset])
}

func (eb *CommandLine) RuneUnderCursor() (rune, int) {
	return utf8.DecodeRune(eb.text[eb.cursorByteOffset:])
}

func (eb *CommandLine) RuneBeforeCursor() (rune, int) {
	return utf8.DecodeLastRune(eb.text[:eb.cursorByteOffset])
}

func (eb *CommandLine) MoveCursorOneRuneBackward() {
	if eb.cursorByteOffset == 0 {
		return
	}
	_, size := eb.RuneBeforeCursor()
	eb.MoveCursorTo(eb.cursorByteOffset - size)
}

func (eb *CommandLine) MoveCursorOneRuneForward() {
	if eb.cursorByteOffset == len(eb.text) {
		return
	}
	_, size := eb.RuneUnderCursor()
	eb.MoveCursorTo(eb.cursorByteOffset + size)
}

func (eb *CommandLine) DeleteRuneBackward() {
	if eb.cursorByteOffset == 0 {
		return
	}

	eb.MoveCursorOneRuneBackward()
	_, size := eb.RuneUnderCursor()
	eb.text = byteSliceRemove(eb.text, eb.cursorByteOffset, eb.cursorByteOffset+size)
}

func (eb *CommandLine) DeleteAll() {
	eb.MoveCursorTo(0)
	eb.text = byteSliceRemove(eb.text, 0, len(eb.text))
}

func (eb *CommandLine) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	eb.text = byteSliceInsert(eb.text, eb.cursorByteOffset, buf[:n])
	eb.MoveCursorOneRuneForward()
}

func (eb *CommandLine) Redraw() {
	w, h := termbox.Size()
	termbox.SetCell(0, h-1, ':', termbox.ColorDefault, termbox.ColorDefault)
	eb.Draw(1, h-1, w-1, 1)
	termbox.SetCursor(1+eb.CursorX(), h-1)
}

func (cl *CommandLine) Run(events <-chan termbox.Event) <-chan []string {
	ch := make(chan []string)
	go func() {
		for {
			switch ev := <-events; ev.Type {
			case termbox.EventKey:
				switch ev.Key {
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

// Please, keep in mind that cursor depends on the value of lineCellOffset, which
// is being set on Draw() call, so.. call this method after Draw() one.
func (eb *CommandLine) CursorX() int {
	return eb.cursorCellOffset - eb.lineCellOffset
}
