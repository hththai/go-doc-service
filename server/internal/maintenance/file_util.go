package maintenance

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// scanFileFolder scans the configured file directory for duplicate base-name files
// and copies them into a temp subfolder with a numeric suffix.
func scanFileFolder() error {
	configFilePath := os.Getenv("FILE_BASE_PATH")
	if configFilePath == "" {
		return fmt.Errorf("FILE_BASE_PATH environment variable is not set")
	}

	dirPath, err := getConfigDirectory(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	fileMap, err := buildFileMap(dirPath)
	if err != nil {
		return fmt.Errorf("failed to build file map: %w", err)
	}

	temp := "temp"
	tempDir := filepath.Join(configFilePath, temp)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	for _, paths := range fileMap {
		if len(paths) > 1 {
			if err := processDuplicateFiles(paths, tempDir); err != nil {
				return fmt.Errorf("failed to process duplicate files: %w", err)
			}
		} else {
			fmt.Printf("Single file path found for key: %s\n", paths[0])
		}
	}

	return nil
}

// buildFileMap traverses the specified directory and its subdirectories,
// collecting files into a map where keys are base names (without extensions)
// and values are slices of full file paths.
func buildFileMap(dirPath string) (map[string][]string, error) {
	fileMap := make(map[string][]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			baseName := filepath.Base(path)
			ext := filepath.Ext(baseName)
			key := baseName[:len(baseName)-len(ext)]

			fileMap[key] = append(fileMap[key], path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return fileMap, nil
}

func processDuplicateFiles(paths []string, tempDir string) error {
	for i, path := range paths {
		ext := filepath.Ext(path)
		baseName := filepath.Base(path[:len(path)-len(ext)])
		newFileName := fmt.Sprintf("%s_%d%s", baseName, i+1, ext)
		newFilePath := filepath.Join(tempDir, newFileName)

		if err := copyFile(path, newFilePath); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	return nil
}

// getConfigDirectory reads the config file to get the directory path.
func getConfigDirectory(configPath string) (string, error) {
	// Implement reading config file and extracting directory path
	// This is a placeholder implementation
	return configPath, nil // Replace with actual logic
}

// copyFile copies a source file to a destination file.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return err
	}

	return nil
}
