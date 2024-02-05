package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/otaviohenrique/zamorak/pkg/engine"
)

func main() {
	// programData, err := os.ReadFile("IBMLogo.ch8")
	// if err != nil {
	// 	fmt.Print(err)
	// }

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")

	if err := ebiten.RunGame(engine.NewRuntime(64, 32)); err != nil {
		log.Fatal(err)
	}
}
