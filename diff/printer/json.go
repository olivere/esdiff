package printer

import (
	"encoding/json"
	"io"

	"github.com/olivere/esdiff/diff"
)

// JSONPrinter prints diffs as JSON, making it easily parseable
// for tools like jq or jiq.
type JSONPrinter struct {
	w         io.Writer
	enc       *json.Encoder
	unchanged bool
	updated   bool
	created   bool
	deleted   bool
}

// NewJSONPrinter creates a new JSONPrinter.
func NewJSONPrinter(w io.Writer, unchanged, updated, created, deleted bool) *JSONPrinter {
	return &JSONPrinter{
		w:         w,
		enc:       json.NewEncoder(w),
		unchanged: unchanged,
		updated:   updated,
		created:   created,
		deleted:   deleted,
	}
}

// Print prints a diff as JSON.
func (p *JSONPrinter) Print(d diff.Diff) error {
	type rowType struct {
		Mode string      `json:"mode"`
		ID   string      `json:"_id"`
		Src  interface{} `json:"src,omitempty"`
		Dst  interface{} `json:"dst,omitempty"`
		// Diff interface{} `json:"diff,omitempty"`
	}

	ok := false

	row := rowType{
		Src: d.Src,
		Dst: d.Dst,
	}

	switch d.Mode {
	case diff.Unchanged:
		row.Mode = "unchanged"
		row.ID = d.Src.ID
		ok = p.unchanged
	case diff.Created:
		row.Mode = "created"
		row.ID = d.Dst.ID
		ok = p.created
	case diff.Updated:
		row.Mode = "updated"
		row.ID = d.Src.ID
		ok = p.updated
	case diff.Deleted:
		row.Mode = "deleted"
		row.ID = d.Src.ID
		ok = p.deleted
	}

	if ok {
		return p.enc.Encode(row)
	}
	return nil
}
