package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	switch os.Getenv("LOG_LEVEL") {
	case "debug", "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "info", "INFO":
		log.SetLevel(log.InfoLevel)
	case "warn", "WARN":
		log.SetLevel(log.WarnLevel)
	case "error", "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "fatal", "FATAL":
		log.SetLevel(log.FatalLevel)
	case "panic", "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
	log.Warn("Starting the converterd service...")

	watchFolders()
}

func watchFolders() {
	CHECK_SECONDS, err := strconv.Atoi(os.Getenv("CHECK_SECONDS"))
	if err != nil || CHECK_SECONDS <= 0 {
		CHECK_SECONDS = 60
	}
	check_frequency := time.Duration(CHECK_SECONDS) * time.Second

	folders_str := os.Getenv("WATCH_FOLDERS")
	if folders_str == "" {
		log.Fatal("No folders specified in WATCH_FOLDERS environment variable")
		return
	}
	folders := strings.Split(folders_str, ",")
	log.Infof("Watching the following folders: %v", folders)

	// Poll the folders every 60 seconds
	ticker := time.NewTicker(check_frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Debug("Checking for new files...")
			// Check for new files in the folders
			for _, folder := range folders {
				log.Debugf("Checking folder: %s", folder)
				checkForNewFiles(folder)
			}
		}
	}
}

func checkForNewFiles(folder string) {
	// Iterate through every file in this folder
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Errorf("Error reading folder %s: %v", folder, err)
		return
	}
	for _, file := range files {
		filePath := folder + "/" + file.Name()
		if file.IsDir() {
			// Walk subdir
			log.Debugf("Found directory: %s", filePath)
			checkForNewFiles(filePath)
			continue
		}
		log.Debugf("Found file: %s", filePath)
		if isNewFile(filePath) {
			processFile(folder, file)
		}
	}
}

func isNewFile(filePath string) bool {
	// If the file is ".flac"
	if strings.HasSuffix(filePath, ".flac") {
		log.Debugf("File is a .flac file: %s", filePath)
		// If the same file name does not exist with ".mp3"
		mp3FileName := strings.TrimSuffix(filePath, ".flac") + ".mp3"
		log.Debugf("Checking for .mp3 file: %s", mp3FileName)
		if _, err := os.Stat(mp3FileName); os.IsNotExist(err) {
			log.Infof("File does not exist with .mp3 extension, converting it: %s", mp3FileName)
			return true
		}
		log.Debugf("File already exists with .mp3 extension: %s", mp3FileName)
	}
	return false
}

func processFile(folder string, file os.FileInfo) {
	inputFile := folder + "/" + file.Name()
	outputFile := folder + "/" + strings.TrimSuffix(file.Name(), ".flac") + ".mp3"
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-codec:a", "libmp3lame", "-b:a", "192k", outputFile)
	log.Infof("Running command: %s", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error converting file %s to %s: %v", inputFile, outputFile, err)
		return
	}
	log.Infof("Successfully converted file %s to %s", inputFile, outputFile)
}
