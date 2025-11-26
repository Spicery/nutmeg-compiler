package bundler

import (
	"encoding/json"
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// // EntryPoint represents a top-level entry point in the bundle.
// type EntryPoint struct {
// 	IdName string `gorm:"primaryKey"`
// }

// // DependsOn represents a dependency relationship between identifiers.
// type DependsOn struct {
// 	IdName string `gorm:"primaryKey;index"`
// 	Needs  string `gorm:"primaryKey;index"`
// }

// // Binding represents a value binding in the bundle.
// type Binding struct {
// 	IdName   string `gorm:"primaryKey"`
// 	Lazy     bool
// 	Value    string
// 	FileName string
// }

// // SourceFile stores the original source file contents.
// type SourceFile struct {
// 	FileName string `gorm:"primaryKey"`
// 	Contents string
// }

// // Annotation stores metadata about bindings.
// type Annotation struct {
// 	IdName          string `gorm:"primaryKey;index"`
// 	AnnotationKey   string `gorm:"primaryKey"`
// 	AnnotationValue string
// }

// Bundler handles the bundling process.
type Bundler struct {
	db          *gorm.DB
	annotations []struct {
		key   string
		value string
	}
}

// NewBundler creates a new bundler with the given database connection.
func NewBundler(dbPath string) (*Bundler, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Bundler{
		db:          db,
		annotations: make([]struct{ key, value string }, 0),
	}, nil
}

// Migrate performs database migrations.
func (b *Bundler) Migrate() error {
	return Migrate(b.db)
}

// CheckMigration checks if the database schema is up to date.
func (b *Bundler) CheckMigration() (bool, error) {
	return CheckMigration(b.db)
}

// ProcessUnit processes a unit node and adds its contents to the bundle.
func (b *Bundler) ProcessUnit(unit *common.Node) error {
	if unit.Name != common.NameUnit {
		return fmt.Errorf("expected unit node, got %s", unit.Name)
	}

	srcPath := unit.Options[common.OptionSrc]

	// Iterate through children of the unit.
	for _, child := range unit.Children {
		switch child.Name {
		case "annotations":
			// Process annotations and add to accumulating list.
			if err := b.processAnnotations(child); err != nil {
				return fmt.Errorf("failed to process annotations: %w", err)
			}

		case common.NameBind:
			// Process bind node.
			if err := b.processBind(child, srcPath); err != nil {
				return fmt.Errorf("failed to process bind: %w", err)
			}

		default:
			// Ignore other nodes for now.
			fmt.Printf("Ignoring top-level node: %s\n", child.Name)
		}
	}

	return nil
}

// processAnnotations extracts annotations and adds them to the accumulating list.
func (b *Bundler) processAnnotations(annotationsNode *common.Node) error {
	// Each child of the annotations node represents an annotation.
	// The child is typically an <id> node with a name attribute.
	for _, child := range annotationsNode.Children {
		if child.Name == common.NameIdentifier {
			key := child.Options[common.OptionName]
			value := "" // Value is not used at present, as per task description.
			b.annotations = append(b.annotations, struct{ key, value string }{key, value})
		} else {
			fmt.Println("Skipping annotation:", child.Name)
		}
	}

	return nil
}

// processBind processes a bind node and inserts/updates the database.
func (b *Bundler) processBind(bindNode *common.Node, srcPath string) error {
	if len(bindNode.Children) != 2 {
		return fmt.Errorf("bind node must have exactly 2 children")
	}

	// First child must be an <id> node.
	idNode := bindNode.Children[0]
	if idNode.Name != common.NameIdentifier {
		return fmt.Errorf("expected id node, got %s", idNode.Name)
	}

	// Second child must be an <fn> node.
	valueNode := bindNode.Children[1]

	// Extract binding information.
	idName := idNode.Options[common.OptionName]
	lazy := idNode.Options[common.OptionLazy] == "true"

	// Serialize the value node to JSON.
	valueJSON, err := json.Marshal(valueNode)
	if err != nil {
		return fmt.Errorf("failed to serialize value node: %w", err)
	}

	// Prepare filename (use srcPath or NULL).
	fileName := ""
	if srcPath != "" {
		fileName = srcPath
	}

	// Upsert the binding.
	binding := Binding{
		IdName:   idName,
		Lazy:     lazy,
		Value:    string(valueJSON),
		FileName: fileName,
	}

	result := b.db.Save(&binding)
	if result.Error != nil {
		return fmt.Errorf("failed to save binding: %w", result.Error)
	}

	// Process accumulated annotations.
	for _, ann := range b.annotations {
		annotation := Annotation{
			IdName:          idName,
			AnnotationKey:   ann.key,
			AnnotationValue: ann.value,
		}
		result := b.db.Save(&annotation)
		if result.Error != nil {
			return fmt.Errorf("failed to save annotation: %w", result.Error)
		}
	}

	// Clear annotations after processing.
	b.annotations = b.annotations[:0]

	return nil
}

// Close closes the database connection.
func (b *Bundler) Close() error {
	sqlDB, err := b.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
