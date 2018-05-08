package diff

import (
	"context"

	"github.com/google/go-cmp/cmp"
)

// Document is a generic document retrieved from Elasticsearch.
type Document struct {
	ID     string                 `json:"_id,omitempty"`
	Source map[string]interface{} `json:"_source,omitempty"`
}

// Mode describes the outcome of comparing two documents.
type Mode int

const (
	// Unchanged means that a document has not been changed between
	// source and destination index.
	Unchanged Mode = iota
	// Created means that a document has been added to the destination
	// that didn't exist in the source index.
	Created
	// Updated means that a document has been found both in the source
	// and destination index, but its contents (_source) has changed.
	Updated
	// Deleted means that a document has been found in the source index
	// but it doesn't exist in the destination index.
	Deleted
)

// Mode returns a string represenation for a mode.
func (m Mode) String() string {
	switch m {
	case Unchanged:
		return "Unchanged"
	case Created:
		return "Created"
	case Updated:
		return "Updated"
	case Deleted:
		return "Deleted"
	default:
		return "<unspecified>"
	}
}

// Diff is the outcome of comparing two documents in source and
// destination index.
type Diff struct {
	Mode Mode
	Src  *Document
	Dst  *Document
}

// Differ compares the documents in the source index to those in
// the destination index. It returns the outcomes via a Diff structure,
// one by one.
func Differ(
	ctx context.Context,
	srcCh <-chan *Document,
	dstCh <-chan *Document,
) (<-chan Diff, <-chan error) {
	diffCh := make(chan Diff)
	errCh := make(chan error)

	go func() {
		defer func() {
			close(diffCh)
			close(errCh)
		}()

		// Both src and dst are nil => no diffs
		if srcCh == nil && dstCh == nil {
			return
		}

		// No src => return all from dst as Created
		if srcCh == nil && dstCh != nil {
			for {
				select {
				case doc, ok := <-dstCh:
					if !ok {
						return
					}
					diffCh <- Diff{Mode: Created, Dst: doc}
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}

		// No dst => return all from src as Deleted
		if srcCh != nil && dstCh == nil {
			for {
				select {
				case doc, ok := <-srcCh:
					if !ok {
						return
					}
					diffCh <- Diff{Mode: Deleted, Src: doc}
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}

		// Read first document from both channels
		srcDoc, dstDoc := <-srcCh, <-dstCh

		// Main loop
		for {
			// Stop early because context might be canceled.
			select {
			default:
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}

			// No more documents from the channels => done
			if srcDoc == nil && dstDoc == nil {
				break
			}

			// No more from dst => everything in src has to be deleted
			if srcDoc != nil && dstDoc == nil {
				diffCh <- Diff{Mode: Deleted, Src: srcDoc}
				for {
					select {
					case doc, ok := <-dstCh:
						if !ok {
							return
						}
						diffCh <- Diff{Mode: Deleted, Src: doc}
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					}
				}
			}

			// No more from src => everything in dst has to be created
			if srcDoc == nil && dstDoc != nil {
				diffCh <- Diff{Mode: Created, Dst: dstDoc}
				for {
					select {
					case doc, ok := <-dstCh:
						if !ok {
							return
						}
						diffCh <- Diff{Mode: Created, Dst: doc}
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					}
				}
			}

			// We have two to compare
			if srcDoc.ID > dstDoc.ID {
				diffCh <- Diff{Mode: Created, Dst: dstDoc}
				dstDoc = nil
				var stop bool
				for !stop {
					select {
					case doc, ok := <-dstCh:
						if !ok {
							stop = true
							break
						}
						dstDoc = doc
						if srcDoc.ID <= dstDoc.ID {
							stop = true
							break
						}
						diffCh <- Diff{Mode: Created, Dst: dstDoc}
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					}
				}
			} else if srcDoc.ID < dstDoc.ID {
				diffCh <- Diff{Mode: Deleted, Src: srcDoc}
				srcDoc = nil
				var stop bool
				for !stop {
					select {
					case doc, ok := <-srcCh:
						if !ok {
							stop = true
							break
						}
						srcDoc = doc
						if srcDoc.ID >= dstDoc.ID {
							stop = true
							break
						}
						diffCh <- Diff{Mode: Deleted, Src: srcDoc}
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					}
				}
			} else {
				// srcDoc.ID == dstDoc.ID
				if cmp.Equal(srcDoc.Source, dstDoc.Source) {
					diffCh <- Diff{Mode: Unchanged, Src: srcDoc, Dst: dstDoc}
				} else {
					diffCh <- Diff{Mode: Updated, Src: srcDoc, Dst: dstDoc}
				}
				srcDoc, dstDoc = <-srcCh, <-dstCh
			}
		}
	}()

	return diffCh, errCh
}
