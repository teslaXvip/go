package distributed

import (
	"log"
	"os"
)

func CheckError(err error) {
	if err != nil {
		log.Printf("erros: %s", err)
		os.Exit(1)
	}
}
