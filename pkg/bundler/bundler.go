package bundler

import (
	"encoding/json"
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	lazy := bindNode.Options[common.OptionLazy] == "true"

	// Convert the value node to JSON.
	var valueJSON []byte

	if valueNode.Name == common.NameFn {
		// Convert <fn> node to FunctionObject.
		funcObj, err := ConvertFnToFunctionObject(valueNode)
		if err != nil {
			return fmt.Errorf("failed to convert function: %w", err)
		}
		valueJSON, err = json.Marshal(funcObj)
		if err != nil {
			return fmt.Errorf("failed to serialize function object: %w", err)
		}
	} else {
		// For non-function values, serialize the node directly.
		var err error
		valueJSON, err = json.Marshal(valueNode)
		if err != nil {
			return fmt.Errorf("failed to serialize value node: %w", err)
		}

	}

	// Find all references to variables in the function body.
	idrefs := findIdentifierReferences(valueNode)

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

	// Upsert the depends-on relationships.
	for _, refName := range idrefs {
		dependency := DependsOn{
			IdName: idName,
			Needs:  refName,
		}
		result := b.db.Save(&dependency)
		if result.Error != nil {
			return fmt.Errorf("failed to save dependency relationship: %w", result.Error)
		}
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

		// If annotation key is "main", create an entry point.
		if ann.key == "main" {
			entryPoint := EntryPoint{
				IdName: idName,
			}
			result := b.db.Save(&entryPoint)
			if result.Error != nil {
				return fmt.Errorf("failed to save entry point: %w", result.Error)
			}
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

// findIdentifierReferences traverses a code tree and collects all unique identifier
// references. It returns a slice of identifier names that are referenced
// in the given node and its descendants, with no duplicates.
func findIdentifierReferences(node *common.Node) []string {
	seen := make(map[string]bool)
	var references []string
	findIdentifierReferencesRecursive(node, &references, seen)
	return references
}

// findIdentifierReferencesRecursive is a helper function that recursively
// traverses the node tree and collects identifier names, avoiding duplicates.
func findIdentifierReferencesRecursive(node *common.Node, references *[]string, seen map[string]bool) {
	if node == nil {
		return
	}

	// If this is an instruction that references an identifier, record it.
	if name, ok := node.Options[common.OptionName]; ok {
		seen[name] = true
		*references = append(*references, name)
	}

	// Recursively process all children.
	for _, child := range node.Children {
		findIdentifierReferencesRecursive(child, references, seen)
	}
}
