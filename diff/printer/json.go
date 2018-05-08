package printer

import (
	"encoding/json"
	"io"

	"github.com/olivere/esdiff/diff"
)

// JSONPrinter prints diffs as JSON, making it easily parseable
// for tools like jq or jiq.
type JSONPrinter struct {
	w       io.Writer
	enc     *json.Encoder
	include int
	exclude int
}

// NewJSONPrinter creates a new JSONPrinter.
func NewJSONPrinter(w io.Writer, include, exclude int) *JSONPrinter {
	return &JSONPrinter{
		w:       w,
		enc:     json.NewEncoder(w),
		include: include,
		exclude: exclude,
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

	row := rowType{
		Src: d.Src,
		Dst: d.Dst,
	}

	switch d.Mode {
	case diff.Unchanged:
		row.Mode = "unchanged"
		row.ID = d.Src.ID
	case diff.Created:
		row.Mode = "created"
		row.ID = d.Dst.ID
	case diff.Updated:
		row.Mode = "updated"
		row.ID = d.Src.ID
	case diff.Deleted:
		row.Mode = "deleted"
		row.ID = d.Src.ID
	}

	return p.enc.Encode(row)
}
