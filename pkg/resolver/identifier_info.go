package resolver

import (
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// ScopeType represents the scope level of an identifier.
type ScopeType string

const (
	GlobalScope ScopeType = "global"
	OuterScope  ScopeType = "outer" // Outer local scope.
	InnerScope  ScopeType = "inner" // Inner local scope.
)

// IdentifierInfo holds information about a resolved identifier.
type IdentifierInfo struct {
	Name          string       // The identifier name.
	UniqueID      uint64       // Unique identifier across all scopes.
	DefDynLevel   int          // Dynamic level where defined.
	ScopeType     ScopeType    // The scope level (global, outer, inner).
	IsAssignable  bool         // Whether this identifier can be assigned to.
	IsConst       bool         // Whether this is a const binding.
	IsProtected   bool         // Whether this identifier can be shadowed.
	LastReference *common.Node // The position of the last reference in the AST traversal.
	DefiningScope *Scope       // The scope where this identifier is defined.
}

func (info *IdentifierInfo) toNode(stype ScopeType) *common.Node {
	return &common.Node{
		Name:     "id",
		Children: []*common.Node{},
		Options: map[string]string{
			common.OptionName:     info.Name,
			common.OptionSerialNo: fmt.Sprintf("%d", info.UniqueID),
			common.OptionScope:    string(stype),
			common.OptionVar:      fmt.Sprintf("%t", info.IsAssignable),
			common.OptionConst:    fmt.Sprintf("%t", info.IsConst),
		},
	}
}
