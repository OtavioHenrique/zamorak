package cmd

import (
	_ "embed"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/otaviohenrique/zamorak/pkg/engine"
	"github.com/otaviohenrique/zamorak/pkg/interpreter"
	"github.com/otaviohenrique/zamorak/pkg/logger"
	"github.com/spf13/cobra"
)

var GameSound []byte

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a CHIP-8 program",
	Long: `Run a CHIP-8 program. Ex:

zamorak run /path/to/rom`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		logLevel, _ := cmd.Flags().GetString("log-level")

		programData, err := os.ReadFile(filePath)

		if err != nil {
			panic(err)
		}

		log := logger.NewLogger(logLevel)

		runtime := engine.NewRuntime(64, 32, GameSound, log)

		inter := interpreter.NewChip8(log)

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

	runCmd.Flags().StringP("log-level", "l", "INFO", "Log Level")
}
