package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NethermindEth/aztec-p2p-explorer/build"
)

func versionMain(cmd *cobra.Command, args []string) {
	fmt.Println("Commit:", build.Commit)
	fmt.Println("Date:", build.BuildDate)
}
