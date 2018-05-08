package printer

import (
	"fmt"
	"io"

	"github.com/google/go-cmp/cmp"

	"github.com/olivere/esdiff/diff"
)

// StdPrinter uses a textual description for diffs.
type StdPrinter struct {
	w io.Writer
}

// NewStdPrinter creates a new StdPrinter.
func NewStdPrinter(w io.WriteCloser) *StdPrinter {
	return &StdPrinter{
		w: w,
	}
}

// Print prints a diff in a textual form. It e.g. doesn't print
// a diff for unchanged documents.
func (p *StdPrinter) Print(d diff.Diff) error {
	switch d.Mode {
	case diff.Unchanged:
		// fmt.Fprintf(p.w, "-\t%v\n", cmp.Diff(d.Src, d.Dst))
	case diff.Created:
		fmt.Fprintf(p.w, "Created\t%v\t%v\n", d.Dst.ID, cmp.Diff(d.Src, d.Dst))
	case diff.Updated:
		fmt.Fprintf(p.w, "Updated\t%v\t%v\n", d.Src.ID, cmp.Diff(d.Src, d.Dst))
	case diff.Deleted:
		fmt.Fprintf(p.w, "Deleted\t%v\n", d.Src.ID)
	}
	return nil
}
