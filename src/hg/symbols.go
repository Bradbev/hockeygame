package hg

import (
	"bytes"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteImage    = ebiten.NewImage(3, 3)
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	whiteImage.Fill(color.White)
}

func MakeSymbol(letters string, r float32, clr color.Color) (*ebiten.Image, error) {
	img := ebiten.NewImage(int(r*2), int(r*2))
	img.Fill(color.Transparent)
	var path vector.Path

	path.Arc(r, r, r, 0, 360, vector.Clockwise)
	path.Close()

	cr, cg, cb, _ := clr.RGBA()
	vertices, indices := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vertices {
		vertices[i].SrcX = 1
		vertices[i].SrcY = 1
		vertices[i].ColorR = float32(cr) / float32(0xffff)
		vertices[i].ColorG = float32(cg) / float32(0xffff)
		vertices[i].ColorB = float32(cb) / float32(0xffff)
		vertices[i].ColorA = 1
	}

	op := &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
		FillRule:  ebiten.FillRuleNonZero,
	}
	img.DrawTriangles(vertices, indices, whiteSubImage, op)

	// Draw the text
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return img, err
	}
	face := &text.GoTextFace{
		Source: s,
		Size:   float64(r),
	}
	w, h := text.Measure(letters, face, 0)
	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(float64(r)-w/2, float64(r)-h/2)
	textOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(img, letters, face, textOp)

	return img, nil
}
