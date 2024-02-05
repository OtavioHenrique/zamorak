package engine

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	colorWhite = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	colorBlack = color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xFF}
)

type Runtime struct {
	width  int
	height int
	keys   []ebiten.Key
	image  *image.RGBA
}

func NewRuntime(width, height int) *Runtime {
	r := new(Runtime)

	r.width = width
	r.height = height

	return r
}

func (g *Runtime) IsPixelSet(col int, row int) bool {

	isSet := g.image.RGBAAt(col, row) == colorWhite

	// if pixel is already "on", we turn off the pixel.
	return isSet
}

func (g *Runtime) Set(col int, row int, on bool) {
	if on {
		g.image.Set(col, row, colorWhite)
	} else {
		g.image.Set(col, row, colorBlack)

		// draw a "shadow" pixel where the previously lit pixel was
		//g.ghostImage.Set(col, row, colorWhite)
	}
}

func (g *Runtime) ClearScreen() {
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			g.image.Set(x, y, color.Black)
		}
	}
}

func (r *Runtime) Update() error {
	r.keys = inpututil.AppendPressedKeys(r.keys[:0])

	return nil
}

func (r *Runtime) Draw(screen *ebiten.Image) {
	screen.WritePixels(r.image.Pix)
	//screen.DrawImage(r.ghostImage, &ebiten.DrawImageOptions{})
}

func (r *Runtime) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return r.width, r.height
}

func ByteToKey(b byte) ebiten.Key {
	var out ebiten.Key
	switch b {
	case 0x0:
		return ebiten.KeyDigit1
	case 0x1:
		return ebiten.KeyDigit2
	case 0x2:
		return ebiten.KeyDigit3
	case 0x3:
		return ebiten.KeyDigit4
	case 0x4:
		return ebiten.KeyQ
	case 0x5:
		return ebiten.KeyW
	case 0x6:
		return ebiten.KeyE
	case 0x7:
		return ebiten.KeyR
	case 0x8:
		return ebiten.KeyA
	case 0x9:
		return ebiten.KeyS
	case 0xA:
		return ebiten.KeyD
	case 0xB:
		return ebiten.KeyF
	case 0xC:
		return ebiten.KeyZ
	case 0xD:
		return ebiten.KeyX
	case 0xE:
		return ebiten.KeyC
	case 0xF:
		return ebiten.KeyV
	}
	return out
}
