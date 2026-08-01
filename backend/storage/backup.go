package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func BackupData(sourceFile string) error {
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return nil
	}

	backupDir := filepath.Join(filepath.Dir(sourceFile), "backup")

	err := os.MkdirAll(backupDir, os.ModePerm)
	if err != nil {
		return err
	}

	originalFilename := filepath.Base(sourceFile)

	ext := filepath.Ext(originalFilename)
	nameWithoutExt := originalFilename[0 : len(originalFilename)-len(ext)]


	timestamp := time.Now().Format("20060102_150405")

	newFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
	destinationFile := filepath.Join(backupDir, newFilename)

	return copyFile(sourceFile, destinationFile)
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}

	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}