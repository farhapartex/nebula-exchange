package config

type MigrationConfig struct {
	DatabaseURL      string
	MigrationsSource string
}

func LoadMigrationConfig() (MigrationConfig, error) {
	databaseConfig, err := LoadDatabaseConfig()
	if err != nil {
		return MigrationConfig{}, err
	}
	return MigrationConfig{
		DatabaseURL:      databaseConfig.URL,
		MigrationsSource: "file://" + readString("MIGRATIONS_PATH", "migrations"),
	}, nil
}
