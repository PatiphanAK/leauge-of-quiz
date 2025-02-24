package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Your Fiber application CLI",
	Long:  `Command line interface for managing your Fiber web application`,
}

// Execute เป็นจุดเริ่มต้นของ CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// เพิ่มคำสั่งย่อยที่นี่
	rootCmd.AddCommand(migrateCmd)
}
