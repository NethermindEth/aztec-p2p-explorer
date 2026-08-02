package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dusted-go/logging/prettylog"
	dotenv "github.com/joho/godotenv"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	// Tentatively load .env file
	_ = dotenv.Load()

	cobra.OnInitialize(initConfig)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("aztec-p2p-explorer")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
}

func printBanner() {
	colours := []string{
		"\033[38;5;81m", // Cyan
		"\033[38;5;75m", // Light Blue
		"\033[38;5;69m", // Sky Blue
		"\033[38;5;63m", // Dodger Blue
		"\033[38;5;57m", // Deep Sky Blue
		"\033[38;5;51m", // Cornflower Blue
		"\033[38;5;45m", // Royal Blue
	}
	banner := `
____________________________  ___________              .__                              
\______   \_____  \______   \ \_   _____/__  _________ |  |   ___________   ___________ 
 |     ___//  ____/|     ___/  |    __)_\  \/  /\____ \|  |  /  _ \_  __ \_/ __ \_  __ \
 |    |   /       \|    |      |        \>    < |  |_> >  |_(  <_> )  | \/\  ___/|  | \/
 |____|   \_______ \____|     /_______  /__/\_ \|   __/|____/\____/|__|    \___  >__|   
                  \/                  \/      \/|__|                           \/       
`
	lines := strings.Split(banner, "\n")

	// remove empty lines
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			lines = append(lines[:i], lines[i+1:]...)
			i--
		}
	}

	for i, line := range lines {
		fmt.Printf("%s%s\n", colours[i], line)
	}

	fmt.Println("\033[0m") // Reset
}

func configureLogging(cmd *cobra.Command, _ []string) *slog.Logger {
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		slog.Debug("Error getting debug flag", "error", err)
	}

	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		slog.Debug("Error getting json flag", "error", err)
	}

	var options *slog.HandlerOptions
	if debug {
		options = &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}
	} else {
		options = &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: false,
		}
	}

	var handler slog.Handler

	if json {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			handler = prettylog.NewHandler(options)
		} else {
			handler = slog.NewTextHandler(os.Stdout, options)
		}
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)
	return logger
}
