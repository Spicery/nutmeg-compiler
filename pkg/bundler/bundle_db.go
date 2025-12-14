package bundler

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// EntryPoint represents a top-level entry point in the bundle.
type EntryPoint struct {
	IdName string `gorm:"primaryKey"`
}

// DependsOn represents a dependency relationship between identifiers.
type DependsOn struct {
	IdName string `gorm:"primaryKey;index"`
	Needs  string `gorm:"primaryKey;index"`
}

// Binding represents a value binding in the bundle.
type Binding struct {
	IdName   string `gorm:"primaryKey"`
	Lazy     bool
	Value    string
	FileName string
}

// SourceFile stores the original source file contents.
type SourceFile struct {
	FileName string `gorm:"primaryKey"`
	Contents string
}

// Annotation stores metadata about bindings.
type Annotation struct {
	IdName          string `gorm:"primaryKey;index"`
	AnnotationKey   string `gorm:"primaryKey"`
	AnnotationValue string
}

// getMigrations returns the list of migrations for the bundle database.
func getMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "202511250001",
			Migrate: func(tx *gorm.DB) error {
				// Create initial schema.
				return tx.AutoMigrate(
					&EntryPoint{},
					&DependsOn{},
					&Binding{},
					&SourceFile{},
					&Annotation{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				// Drop all tables.
				return tx.Migrator().DropTable(
					&Annotation{},
					&SourceFile{},
					&Binding{},
					&DependsOn{},
					&EntryPoint{},
				)
			},
		},
	}
}

// Migrate performs database migrations using gormigrate.
func Migrate(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, getMigrations())
	return m.Migrate()
}

// CheckMigration checks if the database schema is up to date.
func CheckMigration(db *gorm.DB) (bool, error) {
	// Try to get the last migration ID that was applied.
	// If the migrations table doesn't exist, gormigrate will return an error.
	// This indicates that no migrations have been run yet.
	// Use a silent logger to avoid spurious warnings on fresh databases.
	var lastMigration string
	err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Table(gormigrate.DefaultOptions.TableName).
		Select("id").
		Order("id DESC").
		Limit(1).
		Scan(&lastMigration).Error

	if err != nil {
		// If there's an error (e.g., table doesn't exist), migrations are not up to date.
		return false, nil
	}

	// Check if all migrations have been applied.
	migrations := getMigrations()
	if len(migrations) == 0 {
		return true, nil
	}

	// The last migration in our list should match the last applied migration.
	expectedLastID := migrations[len(migrations)-1].ID
	return lastMigration == expectedLastID, nil
}
