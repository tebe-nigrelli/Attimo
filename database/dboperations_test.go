package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddRow(t *testing.T) {
	// Set up test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := SetupDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	currentTime := time.Now().Format("02-01-2006")

	// Test cases
	tests := []struct {
		name         string
		categoryName string
		data         RowData
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Valid General row",
			categoryName: "General",
			data: RowData{
				"Opened":   currentTime,
				"Closed":   currentTime,
				"Note":     "Test note",
				"Project":  "Test Project",
				"Location": "Test Location",
				"File":     dbPath, // Using the test db path as a file that exists
			},
			wantErr: false,
		},
		{
			name:         "Valid Contact row",
			categoryName: "Contact",
			data: RowData{
				"Opened": currentTime,
				"Closed": currentTime,
				"Note":   "Test contact note",
				"Email":  "test@example.com",
				"Phone":  "1234567890",
				"File":   dbPath,
			},
			wantErr: false,
		},
		{
			name:         "Invalid email in Contact",
			categoryName: "Contact",
			data: RowData{
				"Opened": currentTime,
				"Closed": currentTime,
				"Note":   "Test contact note",
				"Email":  "not-an-email",
				"Phone":  "1234567890",
				"File":   dbPath,
			},
			wantErr: true,
			errMsg:  "invalid value for column Email",
		},
		{
			name:         "Missing required field in Financial",
			categoryName: "Financial",
			data: RowData{
				"Opened":   currentTime,
				"Closed":   currentTime,
				"Note":     "Test financial note",
				"Location": "Test Location",
				// Missing Cost_EUR
			},
			wantErr: true,
			errMsg:  "missing value for column Cost_EUR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.AddRow(tt.categoryName, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("AddRow() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("AddRow() error = %v, want %v", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("AddRow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify the row was actually added
			var count int64
			result := db.DB.Table(tt.categoryName).Where(tt.data).Count(&count)
			if result.Error != nil {
				t.Errorf("Failed to verify row: %v", result.Error)
			}
			if count != 1 {
				t.Errorf("Expected 1 row, got %d", count)
			}
		})
	}
}

func TestAddRow_NonexistentCategory(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := SetupDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	err = db.AddRow("NonexistentCategory", RowData{"Field": "value"})
	if err == nil {
		t.Error("Expected error when adding row to nonexistent category")
	}
}
