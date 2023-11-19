package main

import (
	"context"
	"fmt"
	"os"

	"github.com/f010/strap/pkg/lib"
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
directory = "db/migrations"

[seeder]
# This is where all the seeder files resides
# Default to a directory called seeders inside the current working directory 
directory = ""
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
	Migration struct {
		Directory string
	}
	Loaded bool
}

var migrationName string

func main() {
	config, err := lib.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file is not found")
			// os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	// var dm *lib.DatabaseManager

	fmt.Println("config.IsCredentialsValid(): ", config.IsCredentialsValid())

	// if config != nil && config.IsCredentialsValid() {
	conn := connectDatabase(config)
	// TODO:
	_, err = lib.NewDatabaseManager(conn, config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// }

	versionCmd := cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("strap version: %v\n", VERSION)
		},
	}

	initCmd := cobra.Command{
		Use:   "init",
		Short: "Initializes a strap project",
		Run:   lib.Init,
	}

	generateMigrationCmd := cobra.Command{
		Use:   "migration:create [flags]",
		Short: "Generates a new migration file",
		Run:   lib.CreateMigrationFile,
	}
	generateMigrationCmd.Flags().StringVarP(&migrationName, "name", "n", "", "name of the migration (required)")
	generateMigrationCmd.MarkFlagRequired("name")

	migrateCmd := cobra.Command{
		Use:   "migration:run [flags]",
		Short: "Runs pending migrations",
		Run:   lib.MigrateUp,
	}

	rootCmd := &cobra.Command{
		Use:   "strap",
		Short: "A  tool for PostgreSQL database migration and seeding",
	}
	rootCmd.AddCommand(
		&versionCmd,
		&initCmd,
		&generateMigrationCmd,
		&migrateCmd,
	)
	if err = rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func connectDatabase(conf *lib.Config) *pgx.Conn {
	if conf == nil || conf.IsCredentialsValid() {
		return nil
	}
	connString := "postgres://" + conf.Database.User + ":" + conf.Database.Password + "@" + conf.Database.Host + ":5432/" + conf.Database.Name
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return conn
}
