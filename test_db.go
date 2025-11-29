package main

import (
	"fmt"
	"os"
	"uploads/internal/database"
)

func main() {
	// Remove the database file if it exists
	os.Remove("test.db")

	// Create a new database instance
	db, err := database.NewDB("test.db")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	// Test inserting a file record
	fmt.Println("Inserting a test file record...")
	err = db.InsertFile("original_test.txt", "stored_test.txt", 1024, "127.0.0.1:12345")
	if err != nil {
		fmt.Printf("Error inserting file: %v\n", err)
		return
	}

	// Test getting the file by stored name
	fmt.Println("Retrieving the file record...")
	record, err := db.GetFileByStoredName("stored_test.txt")
	if err != nil {
		fmt.Printf("Error getting file: %v\n", err)
		return
	}

	fmt.Printf("Retrieved file: %+v\n", record)

	// Test getting all files
	fmt.Println("Getting all files...")
	allFiles, err := db.GetAllFiles()
	if err != nil {
		fmt.Printf("Error getting all files: %v\n", err)
		return
	}

	fmt.Printf("All files: %+v\n", allFiles)

	// Test deleting the file
	fmt.Println("Deleting the file record...")
	err = db.DeleteFile("stored_test.txt")
	if err != nil {
		fmt.Printf("Error deleting file: %v\n", err)
		return
	}

	// Verify the file was deleted
	_, err = db.GetFileByStoredName("stored_test.txt")
	if err != nil {
	fmt.Printf("Successfully confirmed file was deleted: %v\n", err)
	}

	fmt.Println("Database test completed successfully!")
}
