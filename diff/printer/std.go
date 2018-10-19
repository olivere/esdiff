package printer

import (
	"fmt"
	"io"

	"github.com/google/go-cmp/cmp"

	"github.com/olivere/esdiff/diff"
)

// StdPrinter uses a textual description for diffs.
type StdPrinter struct {
	w         io.Writer
	unchanged bool
	updated   bool
	created   bool
	deleted   bool
}

// NewStdPrinter creates a new StdPrinter.
func NewStdPrinter(w io.WriteCloser, unchanged, updated, created, deleted bool) *StdPrinter {
	return &StdPrinter{
		w:         w,
		unchanged: unchanged,
		updated:   updated,
		created:   created,
		deleted:   deleted,
	}
}

// Print prints a diff in a textual form. It e.g. doesn't print
// a diff for unchanged documents.
func (p *StdPrinter) Print(d diff.Diff) error {
	switch d.Mode {
	case diff.Unchanged:
		if p.unchanged {
			fmt.Fprintf(p.w, "Unchanged\t%v\n", cmp.Diff(d.Src, d.Dst))
		}
	case diff.Created:
		if p.created {
			fmt.Fprintf(p.w, "Created\t%v\t%v\n", d.Dst.ID, cmp.Diff(d.Src, d.Dst))
		}
	case diff.Updated:
		if p.updated {
			fmt.Fprintf(p.w, "Updated\t%v\t%v\n", d.Src.ID, cmp.Diff(d.Src, d.Dst))
		}
	case diff.Deleted:
		if p.deleted {
			fmt.Fprintf(p.w, "Deleted\t%v\n", d.Src.ID)
		}
	}
	return nil
}
