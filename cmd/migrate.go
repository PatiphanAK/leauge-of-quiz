package cmd

import (
	"fmt"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database tables",
	Long:  `Automatically create or update database tables based on model definitions`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting database migration...")

		// เริ่มการเชื่อมต่อกับฐานข้อมูล
		models.InitDB()

		// ทำการ migrate
		err := models.MigrateDB()
		if err != nil {
			fmt.Printf("Migration failed: %v\n", err)
			return
		}

		fmt.Println("Migration completed successfully!")
	},
}
