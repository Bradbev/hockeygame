package hg

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MouseController struct {
	activeCount      int
	dropped          bool
	mouseDown        time.Time
	mouseUp          time.Time
	mouseIsDown      bool
	doubleClick      bool
	mx, my           int
	offsetX, offsetY int
}

func (dc *MouseController) IsDoubleClick() bool {
	return dc.doubleClick
}

func (dc *MouseController) Update() {
	downTime := inpututil.MouseButtonPressDuration(ebiten.MouseButtonLeft)
	dc.doubleClick = false
	switch downTime {
	case 0:
		if dc.mouseIsDown {
			if time.Since(dc.mouseUp) < 300*time.Millisecond {
				dc.doubleClick = true
			}
			dc.mouseUp = time.Now()
		}
		dc.mouseIsDown = false
	case 1:
		dc.mouseIsDown = true
		dc.mouseDown = time.Now()
	}
	dc.updateDragActive()
}

// DrawActive should be called to detect if the mouse is dragging
// DragStart can then be checked to see if this is the first frame of dragging
func (dc *MouseController) DragActive() bool {
	return dc.activeCount > 0
}

func (dc *MouseController) updateDragActive() bool {
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

func (dc *MouseController) dragActive() bool {
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

func (dc *MouseController) DragStart() bool {
	return dc.activeCount == 1
}

func (dc *MouseController) Dropped() bool { return dc.dropped }

func (dc *MouseController) Position() (x, y int) {
	return dc.mx - dc.offsetX, dc.my - dc.offsetY
}

func (dc *MouseController) SetOffset(x, y int) (int, int) {
	dc.offsetX = x
	dc.offsetY = y
	return dc.Position()
}
