package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

const VERSION string = "0.01"

const CONFIG string = `
[database]
host = ""
user = ""
password = ""
name = ""

[migration]
# This is where all the migration files resides
# Default to a directory called migrations inside the current working directory
path = ""

[seeder]
# This is where all the seeder files resides
# Default to a directory called seeders inside the current working directory 
path = ""
`

const MIGRATION_TEMPLATE string = `
	-- Write your migrate up statements here

	---- create above / drop below ----

	-- Write your migrate down statements here
`

type Config struct {
	Database struct {
		Host     string
		User     string
		Password string
		Name     string
	}
}

var migrationName string

func main() {
	versionCmd := cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("strap version: %v\n", "11")
		},
	}

	initCmd := cobra.Command{
		Use:   "init",
		Short: "Initializes a strap project",
		Run:   initConfig,
	}

	generateMigrationCmd := cobra.Command{
		Use:   "migration:create [flags]",
		Short: "Generates a new migration file",
		Run:   createMigrationFile,
	}
	generateMigrationCmd.Flags().StringVarP(&migrationName, "name", "n", "", "name of the migration (required)")
	generateMigrationCmd.MarkFlagRequired("name")

	rootCmd := &cobra.Command{
		Use:   "strap",
		Short: "A  tool for PostgreSQL database migration and seeding",
	}
	rootCmd.AddCommand(
		&versionCmd,
		&initCmd,
		&generateMigrationCmd,
	)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Creates the strap config file
func initConfig(c *cobra.Command, args []string) {
	// TODO: if the config file already exists: warn, accept a flag to override the config file contents

	// TODO?: add the config file to .gitignore

	file, err := os.Create("strap.toml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(CONFIG)
	if err != nil {
		panic(err)
	}
}

func parseConfigFile() *Config {
	f := "strap.toml"
	if _, err := os.Stat(f); err != nil {
		f = "strap.toml"
	}

	var conf Config
	_, err := toml.DecodeFile(f, &conf)
	if err != nil {
		fmt.Printf("%s \n", err.Error())
		panic(err)
	}
	return &conf
}

func tableExists(pool *pgx.Conn, tableName string) (bool, error) {
	var exists bool
	err := pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
		tableName).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func migrationStatus() {
	// TODO
}

// Generates migration file
func createMigrationFile(c *cobra.Command, args []string) {
	var name string
	flag := c.Flag("name")
	name = flag.Value.String()
	if name == "" {
		fmt.Printf("%s", "no name provided")
	}

	var fileName string
	time := time.Now()
	unixTimestamp := time.UnixMilli()

	fileName = fmt.Sprint(unixTimestamp) + "-" + name + ".sql"
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(MIGRATION_TEMPLATE)
	if err != nil {
		panic(err)
	}
}

func createMigrationsTable() {
	conf := parseConfigFile()
	connString := "postgres://" + conf.Database.User + ":" + conf.Database.Password + "@" + conf.Database.Host + ":5432/" + conf.Database.Name
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	// TODO: migrations table name should be configurable through config file
	exists, err := tableExists(conn, "migrations")
	if err != nil {
		fmt.Println(err)
	}

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
			fmt.Println("er", err)
		}
	}
	defer conn.Close(context.Background())
}
