package rand

import (
	"fmt"
	"log"
	"os"

	"github.com/alexsasharegan/dotenv"
)

func CheckForEnvFile () {
	envFile, err := os.Stat("./.env")
	if err == nil {
		err = dotenv.Load()
		fmt.Println(".env file found", envFile)
		if err != nil {
			log.Fatal("No .env file found")
		}
	}
}