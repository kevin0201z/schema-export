package exporter

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// ValidateSplitFileNames checks the entire batch before an exporter writes files.
// Use the same portable comparison on every OS so copying an export to macOS
// cannot silently merge distinct objects. Existing filenames remain unchanged.
func ValidateSplitFileNames(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options ExportOptions) error {
	seen := make(map[string]string)
	fold := cases.Fold()
	add := func(name, kind, suffix string) error {
		if name == "" || strings.ContainsAny(name, "/\\\x00") {
			return fmt.Errorf("invalid split file name for %s %q; use single-file export (omit --split)", kind, name)
		}
		key := norm.NFC.String(fold.String(norm.NFC.String(name + suffix)))
		object := fmt.Sprintf("%s %q", kind, name)
		if previous, exists := seen[key]; exists {
			return fmt.Errorf("split file name collision between %s and %s; use single-file export (omit --split)", previous, object)
		}
		seen[key] = object
		return nil
	}

	for _, table := range tables {
		if err := add(table.Name, "table", ""); err != nil {
			return err
		}
	}
	if options.IncludeViews {
		for _, view := range views {
			if err := add(view.Name, "view", "_view"); err != nil {
				return err
			}
		}
	}
	if options.IncludeProcedures {
		for _, procedure := range procedures {
			if err := add(procedure.Name, "procedure", "_procedure"); err != nil {
				return err
			}
		}
	}
	if options.IncludeFunctions {
		for _, function := range functions {
			if err := add(function.Name, "function", "_function"); err != nil {
				return err
			}
		}
	}
	if options.IncludeTriggers {
		for _, trigger := range triggers {
			if err := add(trigger.Name, "trigger", "_trigger"); err != nil {
				return err
			}
		}
	}
	if options.IncludeSequences {
		for _, sequence := range sequences {
			if err := add(sequence.Name, "sequence", "_sequence"); err != nil {
				return err
			}
		}
	}
	return nil
}
