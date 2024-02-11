package cmd

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/otaviohenrique/zamorak/pkg/engine"
	"github.com/otaviohenrique/zamorak/pkg/interpreter"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a CHIP-8 program",
	Long: `Run a CHIP-8 program. Ex:

zamorak run /path/to/instruction`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		programData, err := os.ReadFile(filePath)

		if err != nil {
			panic(err)
		}

		runtime := engine.NewRuntime(64, 32)

		inter := interpreter.NewChip8()

		go inter.Interpret(runtime, programData)

		ebiten.SetWindowSize(640, 480)
		ebiten.SetWindowTitle("Hello, CHIP-8!")

		if err := ebiten.RunGame(runtime); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
