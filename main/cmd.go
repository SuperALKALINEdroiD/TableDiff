package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var supportedDBs = map[string]bool{
	"mysql": true,
}

type ColumnInfo struct {
	Name    string
	Type    string
	Default string
}

func getConnection(dsn, dbType string) *sql.DB {
	// Trim whitespace and quotes from dbType
	dbType = strings.TrimSpace(dbType)
	dbType = strings.Trim(dbType, "'\"")

	if !supportedDBs[dbType] {
		log.Fatalf("UNSUPPORTED DB TYPE: '%s' (received length: %d, supported: mysql)", dbType, len(dbType))
	}

	if dsn == "" {
		log.Fatalln("CONNECTION STRING CANNOT BE EMPTY")
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		log.Fatalf("failed to connect to %v: %v", dsn, err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping %v: %v", dsn, err)
	}

	return db
}

func getColumnInfo(db *sql.DB, table string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM `%s`", table)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying columns for %s: %w", table, err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var field, colType, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &defaultVal, &extra); err != nil {
			return nil, fmt.Errorf("scanning column: %w", err)
		}
		if field.Valid && colType.Valid {
			defaultValStr := ""
			if defaultVal.Valid {
				defaultValStr = defaultVal.String
			}
			cols = append(cols, ColumnInfo{
				Name:    field.String,
				Type:    colType.String,
				Default: defaultValStr,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating columns: %w", err)
	}

	return cols, nil
}

func main() {
	leftDSN := os.Getenv("LEFT_DB_DSN")
	rightDSN := os.Getenv("RIGHT_DB_DSN")
	leftType := os.Getenv("LEFT_DB_TYPE")
	rightType := os.Getenv("RIGHT_DB_TYPE")
	leftTableName := os.Getenv("LEFT_TABLE")
	rightTableName := os.Getenv("RIGHT_TABLE")

	if leftDSN == "" {
		log.Fatal("LEFT_DB_DSN environment variable is required")
	}
	if rightDSN == "" {
		log.Fatal("RIGHT_DB_DSN environment variable is required")
	}
	if leftType == "" {
		log.Fatal("LEFT_DB_TYPE environment variable is required")
	}
	if rightType == "" {
		log.Fatal("RIGHT_DB_TYPE environment variable is required")
	}
	if leftTableName == "" {
		log.Fatal("LEFT_TABLE environment variable is required")
	}
	if rightTableName == "" {
		log.Fatal("RIGHT_TABLE environment variable is required")
	}

	leftConn := getConnection(leftDSN, leftType)
	rightConn := getConnection(rightDSN, rightType)

	var wg sync.WaitGroup
	var leftCols []ColumnInfo
	var rightCols []ColumnInfo
	var leftErr, rightErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		leftCols, leftErr = getColumnInfo(leftConn, leftTableName)
	}()

	go func() {
		defer wg.Done()
		rightCols, rightErr = getColumnInfo(rightConn, rightTableName)
	}()

	wg.Wait()

	if leftErr != nil {
		log.Fatal(leftErr)
	}
	if rightErr != nil {
		log.Fatal(rightErr)
	}

	log.Printf("Left table %s columns: %d\n", leftTableName, len(leftCols))
	log.Printf("Right table %s columns: %d\n", rightTableName, len(rightCols))

	diffColumns(leftCols, rightCols)

	leftOrderBy := os.Getenv("LEFT_ORDER_BY")
	if leftOrderBy == "" {
		leftOrderBy = "id"
	}
	rightOrderBy := os.Getenv("RIGHT_ORDER_BY")
	if rightOrderBy == "" {
		rightOrderBy = "id"
	}

	defer leftConn.Close()
	defer rightConn.Close()

}
