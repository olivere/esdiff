package printer

import (
	"github.com/olivere/esdiff/diff"
)

// Printer prints a diff using a specific output format, e.g. JSON.
type Printer interface {
	Print(diff.Diff) error
}
