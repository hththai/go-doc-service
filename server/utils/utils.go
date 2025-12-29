package utils

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func IsProduction() bool {
	return os.Getenv("GO_ENV") == "production"
}

// TODO: May need to make it more automation.
func GetConfigJWT() string {
	_ = godotenv.Load(".env")

	env := os.Getenv("GO_ENV")
	switch env {
	case "development":
		_ = godotenv.Overload(".env.development")
	case "production":
		_ = godotenv.Overload(".env.production")
	default:
		log.Println("GO_ENV not set, using defaults from .env")
	}
	return os.Getenv("JWT_SECRET")
}

// Custom Util function to create a folder and name if it is not exist.
func createFolderAndFile(folderName string, fileWithExt string) (*os.File, error) {

	err := os.MkdirAll("./"+folderName, os.ModePerm)
	if err != nil {

		return nil, errors.New("failed to create dictionary")
	}

	// Open and Create the log file.
	logFile, err := os.OpenFile("./"+folderName+"/"+fileWithExt, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

		return nil, errors.New("failed to create file")
	}

	return logFile, nil
}
