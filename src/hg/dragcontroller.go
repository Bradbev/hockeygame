package hg

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type DragController struct {
	activeCount      int
	dropped          bool
	mouseDown        time.Time
	mx, my           int
	offsetX, offsetY int
}

func (dc *DragController) DragActive() bool {
	ret := dc.dragActive()
	dc.dropped = false
	if ret {
		dc.activeCount++
	} else {
		dc.dropped = dc.activeCount > 0
		dc.activeCount = 0
		dc.SetOffset(0, 0)
	}
	return ret
}

func (dc *DragController) dragActive() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		dc.mouseDown = time.Now()
		dc.mx, dc.my = ebiten.CursorPosition()
		return false
	}
	if inpututil.MouseButtonPressDuration(ebiten.MouseButtonLeft) > 0 {
		x, y := ebiten.CursorPosition()
		if dc.mx != x || dc.my != y {
			dc.mx, dc.my = x, y
			return true
		}
		if time.Since(dc.mouseDown) > 100*time.Millisecond {
			return true
		}
	}
	// todo - touch handling
	return false
}

func (dc *DragController) DragStart() bool {
	return dc.activeCount == 1
}

func (dc *DragController) Dropped() bool { return dc.dropped }

func (dc *DragController) Position() (x, y int) {
	return dc.mx - dc.offsetX, dc.my - dc.offsetY
}

func (dc *DragController) SetOffset(x, y int) (int, int) {
	dc.offsetX = x
	dc.offsetY = y
	return dc.Position()
}
