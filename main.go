package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create upload directory
	uploadDir := "/tmp/uploads"
	os.MkdirAll(uploadDir, 0755)

	// Save file
	filePath := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	// Get absolute path for LMD
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Failed to get absolute path", http.StatusInternalServerError)
		return
	}

	// Scan with LMD
	fmt.Println(absPath)

	scanCmd := exec.Command("maldet", "--scan-file", absPath)
	output, err := scanCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Scan error: %v\n%s", err, output), http.StatusInternalServerError)
		return
	}

	// Check result properly

	if !strings.Contains(string(output), "Hits: 0") {
		os.Remove(filePath)
		fmt.Fprintf(w, "File is infected and has been removed.\n%s", output)
		return
	}

	fmt.Fprintf(w, "File uploaded and scanned successfully!\n%s", output)
}

func isInfected(scanOutput string) bool {
	// LMD summary usually contains "0 malware hits" if clean
	return !strings.Contains(scanOutput, "0 malware hits")
}

func main() {

	// Create timestamped folder
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	cwd, _ := os.Getwd()
	logDir := filepath.Join(cwd, "logs", timestamp)

	// Ensure folder exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	// Log file path inside timestamped folder
	logFile := filepath.Join(logDir, "myapp.log")

	// Configure Logrus with lumberjack
	log := logrus.New()
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,  // Keep 5 rotated files
		MaxAge:     7,  // Days to keep old logs
		Compress:   true,
	})

	// Set log level to Debug so Debug messages are shown
	log.SetLevel(logrus.DebugLevel)

	// Example logs
	log.Info("Application started with timestamped folder and lumberjack rotation")
	log.Debug("This is a debug message for troubleshooting")
	log.Warn("This is a warning message")
	log.Error("This is an error message")

	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
