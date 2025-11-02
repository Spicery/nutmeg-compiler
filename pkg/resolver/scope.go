package resolver

import (
	"fmt"

	"github.com/spicery/nutmeg-compiler/pkg/common"
)

// Scope represents a single scope level in the scope stack.
type Scope struct {
	Level        int                        // Nesting level (0 = global).
	DynamicLevel int                        // Nesting level counting only dynamic scopes (0 = global).
	Identifiers  map[string]*IdentifierInfo // Maps identifier names to their metadata.
	Parent       *Scope                     // Parent scope for lookups.
	IsDynamic    bool                       // True if this is a dynamic scope (def, fn), false if lexical (if, for, let).
	Node         *common.Node               // The AST node that introduced this scope.
	Captured     map[uint64]*IdentifierInfo // Identifiers captured from outer scopes.
}

// NewChildScope creates a new child scope of the current scope.
func (s *Scope) NewChildScope(isDynamic bool, node *common.Node) *Scope {
	dynamicLevel := s.DynamicLevel
	if isDynamic {
		dynamicLevel++
	}
	scope := &Scope{
		Level:        s.Level + 1,
		DynamicLevel: dynamicLevel,
		Identifiers:  make(map[string]*IdentifierInfo),
		Parent:       s,
		IsDynamic:    isDynamic,
		Node:         node,
	}
	if isDynamic {
		scope.Captured = make(map[uint64]*IdentifierInfo)
	}
	return scope
}

func (s *Scope) isClosureScope() bool {
	return s.IsDynamic && len(s.Captured) > 0
}

func (s *Scope) captureIdentifier(info *IdentifierInfo, r *Resolver) {
	fmt.Println("Capturing in scope level:", s.Level, "dynamic level:", s.DynamicLevel, "identifier:", info.Name)
	if s.Captured[info.UniqueID] == nil {
		if s.Captured == nil {
			s.Captured = make(map[uint64]*IdentifierInfo)
		}
		s.Captured[info.UniqueID] = info
		r.Closures = append(r.Closures, s)
		fmt.Println("Captured identifier:", info.Name, "in scope level:", s.Level, "dynamic level:", s.DynamicLevel, "isClosure:", s.isClosureScope())
	}
}

// getInitialScopeType returns the scope type of the current scope.
func (s *Scope) getInitialScopeType() ScopeType {
	switch s.Level {
	case 0:
		return GlobalScope
	default:
		return InnerScope
	}
}

func (s *Scope) NewIdentifierInfo(name string, uniqueID uint64) *IdentifierInfo {
	info := &IdentifierInfo{
		Name:          name,
		UniqueID:      uniqueID,
		DefDynLevel:   s.DynamicLevel,
		ScopeType:     s.getInitialScopeType(),
		IsAssignable:  false, // TODO: Will be set based on var vs const binding.
		IsConst:       false, // TODO: Will be set based on binding type.
		LastReference: nil,   // Will be updated as identifier is referenced.
		DefiningScope: s,     // Store the scope where this identifier is defined.
	}
	s.Identifiers[name] = info
	return info
}
