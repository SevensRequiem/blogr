package backups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// BackupFiles creates a backup of the files
func BackupFiles() error {
	fmt.Println("Backing up files...")
	dir := "blogs"
	files, err := ListFiles(dir)
	if err != nil {
		fmt.Println("Error listing files:", err)
		return err
	}

	for _, file := range files {
		// Get the relative path of the file
		relPath, err := filepath.Rel(dir, file)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Create the backup directory structure
		backupDir := filepath.Join("backups", filepath.Dir(relPath))
		if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		// Copy the file to the backup location
		backupFilePath := filepath.Join(backupDir, filepath.Base(file))
		if err := copyFile(file, backupFilePath); err != nil {
			return fmt.Errorf("failed to backup file: %s, %v", file, err)
		}
	}

	fmt.Println("Backup complete")
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src string, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dst, input, 0644); err != nil {
		return err
	}
	return nil
}

// ListFiles lists all files in the specified directory and its subdirectories
func ListFiles(dir string) ([]string, error) {
	var fileNames []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() { // Only include files, not directories
			fileNames = append(fileNames, path) // Use full path for clarity
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileNames, nil
}
