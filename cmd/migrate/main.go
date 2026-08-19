package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"habit-tracker-api/internal/config"
	"habit-tracker-api/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Error: укажите имя миграции")
			fmt.Println("Usage: go run cmd/migrate/main.go create <migration_name>")
			os.Exit(1)
		}
		migrationName := os.Args[2]
		createMigration(migrationName)

	case "up":
		runMigrations()

	default:
		fmt.Printf("Error: неизвестная команда '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: go run cmd/migrate/main.go <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  create <name>  - Создать новую миграцию")
	fmt.Println("  up            - Выполнить все миграции")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go create create_tables")
	fmt.Println("  go run cmd/migrate/main.go up")
}

func createMigration(migrationName string) {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	migrationsDir := filepath.Join(workDir, "migrations")

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		fmt.Printf("Error: failed to create migrations directory: %v\n", err)
		os.Exit(1)
	}

	migrationNumber := getNextMigrationNumber(migrationsDir)
	fileName := fmt.Sprintf("%03d_%s.sql", migrationNumber, migrationName)
	filePath := filepath.Join(migrationsDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("Error: migration file %s already exists\n", fileName)
		os.Exit(1)
	}

	template := generateMigrationTemplate(migrationName)

	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		fmt.Printf("Error: failed to create migration file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Migration created: %s\n", fileName)
	fmt.Printf("📝 Edit file: %s\n", filePath)
}

func runMigrations() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := cfg.Database.DSN()
	if err := database.Connect(dsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	migrationsDir := filepath.Join(workDir, "migrations")

	files, err := getMigrationFiles(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No migrations found")
		return
	}

	fmt.Printf("Found %d migration(s)\n", len(files))

	for _, file := range files {
		filePath := filepath.Join(migrationsDir, file)
		fmt.Printf("📦 Running migration: %s\n", file)

		sqlBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		_, err = database.DB.Exec(string(sqlBytes))
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", file, err)
		}

		fmt.Printf("✅ Migration %s completed\n", file)
	}

	fmt.Println("\n✅ All migrations completed successfully!")
}

func getMigrationFiles(migrationsDir string) ([]string, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		numI := extractMigrationNumber(migrations[i])
		numJ := extractMigrationNumber(migrations[j])
		return numI < numJ
	})

	return migrations, nil
}

func extractMigrationNumber(fileName string) int {
	parts := strings.Split(fileName, "_")
	if len(parts) > 0 {
		if num, err := strconv.Atoi(parts[0]); err == nil {
			return num
		}
	}
	return 0
}

func getNextMigrationNumber(migrationsDir string) int {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 1
	}

	maxNumber := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		num := extractMigrationNumber(file.Name())
		if num > maxNumber {
			maxNumber = num
		}
	}

	return maxNumber + 1
}

func generateMigrationTemplate(migrationName string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	template := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- CREATE TABLE IF NOT EXISTS table_name (
--     id BIGSERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

`, migrationName, timestamp)

	return template
}
