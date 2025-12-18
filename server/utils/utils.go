package utils

import (
	"errors"
	"os"
)

func IsProduction() bool {
	return os.Getenv("GO_ENV") == "production"
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
