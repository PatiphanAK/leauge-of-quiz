// main.go
package main

import (
	"github.com/patiphanak/league-of-quiz/cmd"
	models "github.com/patiphanak/league-of-quiz/model"
)

func main() {
	cmd.Execute()

	// Connect to database
	models.InitDB()
}
