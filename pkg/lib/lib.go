package lib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

const CONFIG string = `
host = ""
user = ""
password = ""
name = ""

# This is where all the migration files resides
# Default to a directory called migrations inside the current working directory
migration_directory = "db/migrations"

# This is where all the seeder files resides
# Default to a directory called seeders inside the current working directory 
seeder_directory = ""
`

// Default migration files directory
var migrationsDir string = "db/migrations"

type Config struct {
	Database struct {
		Host     string
		User     string
		Password string
		Name     string
	}
	Migration struct {
		Directory string
	}
	Loaded bool
}

type Migration struct {
	Id        uint
	Name      string
	Timestamp time.Time

	Filename string
	Pending  bool
}

type DatabaseManager struct {
	Config     *Config
	Migrations []*Migration
	Connection *pgx.Conn
}

// type InvalidCredentialsErr struct {
// 	MissingFields []string
// }

// func (e *InvalidCredentialsErr) GetMissingFields() []string {
// 	return e.MissingFields
// }

func NewDatabaseManager(conn *pgx.Conn, conf *Config) (*DatabaseManager, error) {
	if conf != nil && conf.IsCredentialsValid() {
		return nil, errors.New("no database credentials available")
	}

	dm := DatabaseManager{Connection: conn}
	return &dm, nil
}

func (c *Config) IsCredentialsValid() bool {
	if c.Database.Host == "" || c.Database.Password == "" ||
		c.Database.Name == "" || c.Database.User == "" {
		return false
	}
	return true
}

func LoadConfig() (*Config, error) {
	f := "strap.toml"
	if _, err := os.Stat(f); err != nil {
		return nil, err
	}

	var conf Config
	_, err := toml.DecodeFile(f, &conf)
	if err != nil {
		return nil, err
	}
	conf.Loaded = true
	return &conf, nil
}

func Init(c *cobra.Command, args []string) {
	// TODO: If the config file already exists:
	//       Check if there is data in the config file that set by user
	//	     If data: Warn
	//       If data: Accept a flag to override the config file contents

	// TODO?: Add the config file to .gitignore
	// TODO: The config variables should be able take from env files as well, instead of strap's config file
	fmt.Println("creating ")
	file, err := os.Create("strap.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString(CONFIG)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func CreateMigrationFile(c *cobra.Command, args []string) {
	var name string
	flag := c.Flag("name")
	name = flag.Value.String()
	if name == "" {
		fmt.Printf("%s", "no name provided") // TODO: custom error type
		c.Help()
	}

	conf, err := LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Config file not found: Please run 'strap init'")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	if conf.Migration.Directory != "" {
		migrationsDir = conf.Migration.Directory
	}
	path := filepath.Join(".", migrationsDir)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var filePath string
	time := time.Now()
	unixTimestamp := time.UnixMilli()

	// FIX: Use OS specific Separator after the 'directory path' (before the filename)
	filePath = path + "/" + fmt.Sprint(unixTimestamp) + "-" + name + ".sql"

	fmt.Println("filename: ", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString("MIGRATION_TEMPLATE")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
}

func tableExists(conn *pgx.Conn, tableName string) bool {
	var exists bool
	err := conn.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
		tableName).Scan(&exists)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return exists
}

func createMigrationsTable(conn *pgx.Conn) {
	// TODO: migrations table name should be configurable through config file
	exists := tableExists(conn, "migrations")

	if !exists {
		query := `
	    CREATE TABLE migrations (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		timestamp TIMESTAMP DEFAULT current_timestamp
		);
	`
		_, err := conn.Query(context.Background(), query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func findExecutedMigrations(conn *pgx.Conn) []*Migration {
	var rows pgx.Rows
	var migrations []*Migration
	rows, err := conn.Query(context.Background(), `
	SELECT * FROM migrations`,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for rows.Next() {
		var m Migration
		if err = rows.Scan(
			&m.Id,
			&m.Name,
			&m.Timestamp,
		); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		migrations = append(migrations, &m)
	}
	return migrations
}

// Get the list of files inside the 'migrations' directory
func listMigrationFiles(conf *Config) []string {
	if conf.Migration.Directory != "" {
		migrationsDir = conf.Migration.Directory
	}
	path := filepath.Join(".", migrationsDir)

	var files []string
	contents, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, migration := range contents {
		files = append(files, migration.Name())
	}
	return files
}

func (db *DatabaseManager) MigrateUp(c *cobra.Command, args []string) {
	// conn := connectDatabase(&lib.Config{})
	// defer d.Connection.Close(context.Background())
	// Create the migrations table if not exists
	createMigrationsTable(db.Connection)

	// Find completed migrations from migrations table
	executedMigrations := findExecutedMigrations(db.Connection)

	fmt.Println("executed migrations: ", executedMigrations)

	// List the migration files that are available in the migrations directory
	migrationFiles := listMigrationFiles(db.Config)
	fmt.Println("migrationFiles:", migrationFiles)

	if len(migrationFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No migration files found")
		os.Exit(1)
	}

	var pendingMigrations []string

outer:
	for _, f := range migrationFiles {
		for _, m := range executedMigrations {
			if f == m.Name {
				continue outer
			}
		}
		pendingMigrations = append(pendingMigrations, f)
	}
	fmt.Println("pending: ", pendingMigrations)

	// TODO: Run pending migrations
}

func (d *DatabaseManager) MigrateDown(c *cobra.Command, args []string) {
	// TODO
}
