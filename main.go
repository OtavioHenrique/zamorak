package main

import (
	_ "embed"

	"github.com/otaviohenrique/zamorak/cmd"
)

//go:embed resources/sound/song.wav
var soundEffect []byte

func main() {
	cmd.GameSound = soundEffect
	cmd.Execute()
}
