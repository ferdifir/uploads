package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type FileRecord struct {
	ID           int64     `json:"id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	UploadTime   time.Time `json:"upload_time"`
	FileSize     int64     `json:"file_size"`
	UploadAddr   string    `json:"upload_addr"`
}

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	// Create the files table if it doesn't exist
	if err := db.createTable(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

func (db *DB) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_name TEXT NOT NULL,
		stored_name TEXT NOT NULL UNIQUE,
		upload_time DATETIME NOT NULL,
		file_size INTEGER NOT NULL,
		upload_addr TEXT NOT NULL
	);
	`

	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) InsertFile(originalName, storedName string, fileSize int64, uploadAddr string) error {
	query := `
	INSERT INTO files (original_name, stored_name, upload_time, file_size, upload_addr)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query, originalName, storedName, time.Now(), fileSize, uploadAddr)
	return err
}

func (db *DB) GetFileByStoredName(storedName string) (*FileRecord, error) {
	query := `
	SELECT id, original_name, stored_name, upload_time, file_size, upload_addr
	FROM files
	WHERE stored_name = ?
	`

	row := db.conn.QueryRow(query, storedName)

	var record FileRecord
	err := row.Scan(
		&record.ID,
		&record.OriginalName,
		&record.StoredName,
		&record.UploadTime,
		&record.FileSize,
		&record.UploadAddr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found: %s", storedName)
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &record, nil
}

func (db *DB) GetAllFiles() ([]*FileRecord, error) {
	query := `
	SELECT id, original_name, stored_name, upload_time, file_size, upload_addr
	FROM files
	ORDER BY upload_time DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var records []*FileRecord
	for rows.Next() {
		var record FileRecord
		err := rows.Scan(
			&record.ID,
			&record.OriginalName,
			&record.StoredName,
			&record.UploadTime,
			&record.FileSize,
			&record.UploadAddr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

func (db *DB) DeleteFile(storedName string) error {
	query := `
	DELETE FROM files WHERE stored_name = ?
	`

	result, err := db.conn.Exec(query, storedName)
	if err != nil {
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found in database: %s", storedName)
	}

	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
