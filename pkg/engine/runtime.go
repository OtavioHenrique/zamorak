package engine

import (
	"bytes"
	"image"
	"image/color"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	colorWhite = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	colorBlack = color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xFF}
)

type Runtime struct {
	width   int
	height  int
	keys    []ebiten.Key
	image   *image.RGBA
	aPlayer *audio.Player
	logger  *slog.Logger
}

func NewRuntime(width, height int, gameSound []byte, logger *slog.Logger) *Runtime {
	initializedImage := ebiten.NewImage(64, 32)
	initializedImage.Fill(color.RGBA{}) // initialize all pixels to black, 0 alpha.

	r := new(Runtime)

	r.width = width
	r.height = height
	r.image = image.NewRGBA(image.Rect(0, 0, 64, 32))

	decodedSong, err := wav.DecodeWithoutResampling(bytes.NewReader(gameSound))

	if err != nil {
		logger.Error("ERROR DECODING SOUND FILE", "err", err, "bytes", gameSound)

		os.Exit(1)
	}

	audioContext := audio.NewContext(44_000)
	audioPlayer, err := audioContext.NewPlayer(decodedSong)
	if err != nil {
		logger.Error("ERROR CREATING AUDIO PLAYER")

		os.Exit(1)
	}

	r.aPlayer = audioPlayer

	r.ClearScreen()

	return r
}

func (r *Runtime) PlayAudio() {
	if !r.aPlayer.IsPlaying() {
		r.aPlayer.Play()
	}
}

func (r *Runtime) StopAudio() {
	if r.aPlayer.IsPlaying() {
		r.aPlayer.Pause()
		r.aPlayer.Rewind()
	}
}

func (r *Runtime) IsPixelSet(col int, row int) bool {

	isSet := r.image.RGBAAt(col, row) == colorWhite

	return isSet
}

func (r *Runtime) Set(col int, row int, on bool) {
	if on {
		r.image.Set(col, row, colorWhite)
	} else {
		r.image.Set(col, row, colorBlack)
	}
}

func (r *Runtime) ClearScreen() {
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			r.image.Set(x, y, color.Black)
		}
	}
}

func (r *Runtime) Update() error {
	r.keys = inpututil.AppendPressedKeys(r.keys[:0])

	return nil
}

func (r *Runtime) Draw(screen *ebiten.Image) {
	screen.WritePixels(r.image.Pix)
}

func (r *Runtime) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return r.width, r.height
}

func (g *Runtime) WaitKeyPress(c chan byte) {
	for _, key := range g.keys {
		out := keyToByte(key)
		c <- out
	}
}

func keyToByte(key ebiten.Key) byte {
	var out byte
	switch key {
	case ebiten.KeyDigit1:
		out = 0x0
	case ebiten.KeyDigit2:
		out = 0x1
	case ebiten.KeyDigit3:
		out = 0x2
	case ebiten.KeyDigit4:
		out = 0x3
	case ebiten.KeyQ:
		out = 0x4
	case ebiten.KeyW:
		out = 0x5
	case ebiten.KeyE:
		out = 0x6
	case ebiten.KeyR:
		out = 0x7
	case ebiten.KeyA:
		out = 0x8
	case ebiten.KeyS:
		out = 0x9
	case ebiten.KeyD:
		out = 0xA
	case ebiten.KeyF:
		out = 0xB
	case ebiten.KeyZ:
		out = 0xC
	case ebiten.KeyX:
		out = 0xD
	case ebiten.KeyC:
		out = 0xE
	case ebiten.KeyV:
		out = 0xF
	}

	return out
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
