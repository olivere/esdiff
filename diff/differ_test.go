package diff

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/fortytw2/leaktest"
)

var differTests = []struct {
	Srcs, Dsts []*Document
	Errs       []error
	Diffs      []Diff
}{
	// #0
	{
		Srcs:  nil,
		Dsts:  nil,
		Errs:  nil,
		Diffs: nil,
	},
	// #1
	{
		Srcs: nil,
		Dsts: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One"}},
		},
		Errs: nil,
		Diffs: []Diff{
			{
				Mode: Created,
				Src:  nil,
				Dst:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One"}},
			},
		},
	},
	// #2
	{
		Srcs: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One"}},
		},
		Dsts: nil,
		Errs: nil,
		Diffs: []Diff{
			{
				Mode: Deleted,
				Src:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One"}},
				Dst:  nil,
			},
		},
	},
	// #3
	{
		Dsts: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One"}},
			{ID: "2", Source: map[string]interface{}{"Name": "Two"}},
		},
		Srcs: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One"}},
			{ID: "3", Source: map[string]interface{}{"Name": "Three"}},
		},
		Errs: nil,
		Diffs: []Diff{
			{
				Mode: Unchanged,
				Src:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One"}},
				Dst:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One"}},
			},
			{
				Mode: Created,
				Src:  nil,
				Dst:  &Document{ID: "2", Source: map[string]interface{}{"Name": "Two"}},
			},
			{
				Mode: Deleted,
				Src:  &Document{ID: "3", Source: map[string]interface{}{"Name": "Three"}},
				Dst:  nil,
			},
		},
	},
	// #4
	{
		Srcs: []*Document{
			{ID: "2", Source: map[string]interface{}{"Name": "Two"}},
			{ID: "3", Source: map[string]interface{}{"Name": "Three"}},
			{ID: "4", Source: map[string]interface{}{"Name": "Four", "Value": 3}},
			{ID: "5", Source: map[string]interface{}{"Name": "Five"}},
			{ID: "6", Source: map[string]interface{}{"Name": "Six"}},
		},
		Dsts: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One"}},
			{ID: "4", Source: map[string]interface{}{"Name": "Four", "Value": 4}},
			{ID: "6", Source: map[string]interface{}{"Name": "Six"}},
		},
		Errs: nil,
		Diffs: []Diff{
			{
				Mode: Created,
				Src:  nil,
				Dst:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One"}},
			},
			{
				Mode: Deleted,
				Src:  &Document{ID: "2", Source: map[string]interface{}{"Name": "Two"}},
				Dst:  nil,
			},
			{
				Mode: Deleted,
				Src:  &Document{ID: "3", Source: map[string]interface{}{"Name": "Three"}},
				Dst:  nil,
			},
			{
				Mode: Updated,
				Src:  &Document{ID: "4", Source: map[string]interface{}{"Name": "Four", "Value": 3}},
				Dst:  &Document{ID: "4", Source: map[string]interface{}{"Name": "Four", "Value": 4}},
			},
			{
				Mode: Deleted,
				Src:  &Document{ID: "5", Source: map[string]interface{}{"Name": "Five"}},
				Dst:  nil,
			},
			{
				Mode: Unchanged,
				Src:  &Document{ID: "6", Source: map[string]interface{}{"Name": "Six"}},
				Dst:  &Document{ID: "6", Source: map[string]interface{}{"Name": "Six"}},
			},
		},
	},
	// #5
	{
		Srcs: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One", "Price": 599.00}},
			{ID: "3", Source: map[string]interface{}{"Name": "Three", "Price": 2599.00}},
		},
		Dsts: []*Document{
			{ID: "1", Source: map[string]interface{}{"Name": "One", "Price": 599.00}},
			{ID: "2", Source: map[string]interface{}{"Name": "Two", "Price": 2599.00}},
		},
		Errs: nil,
		Diffs: []Diff{
			{
				Mode: Unchanged,
				Src:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One", "Price": 599.00}},
				Dst:  &Document{ID: "1", Source: map[string]interface{}{"Name": "One", "Price": 599.00}},
			},
			{
				Mode: Created,
				Src:  nil,
				Dst:  &Document{ID: "2", Source: map[string]interface{}{"Name": "Two", "Price": 2599.00}},
			},
			{
				Mode: Deleted,
				Src:  &Document{ID: "3", Source: map[string]interface{}{"Name": "Three", "Price": 2599.00}},
				Dst:  nil,
			},
		},
	}}

func TestDiffer(t *testing.T) {
	for i, tt := range differTests {
		ctx := context.Background()
		done := make(chan struct{})
		var errs []error
		var diffs []Diff

		// Generator for src
		srcCh := make(chan *Document)
		go func() {
			defer close(srcCh)
			for _, doc := range tt.Srcs {
				srcCh <- doc
			}
		}()

		// Generator for dst
		dstCh := make(chan *Document)
		go func() {
			defer close(dstCh)
			for _, doc := range tt.Dsts {
				dstCh <- doc
			}
		}()

		// Process diffs
		go func() {
			defer close(done)
			diffCh, errCh := Differ(ctx, srcCh, dstCh)
			var done bool
			for !done {
				select {
				case d, ok := <-diffCh:
					if !ok {
						return
					}
					diffs = append(diffs, d)
				case err, ok := <-errCh:
					if !ok {
						return
					}
					errs = append(errs, err)
				}
			}
		}()

		// Wait until we are done or we get a timeout
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout")
		}
		leaktest.Check(t)()

		if want, have := len(tt.Errs), len(errs); want != have {
			t.Fatalf("#%d: len(Errors): want %d, have %d\n%v", i, want, have, cmp.Diff(tt.Errs, errs))
		}
		for k := 0; k < len(errs); k++ {
			if want, have := tt.Errs[k], errs[k]; want != have {
				t.Fatalf("#%d: Error[%d]: want %v, have %v", i, k, want, have)
			}
		}
		if want, have := len(tt.Diffs), len(diffs); want != have {
			t.Fatalf("#%d: len(Diffs): want %d, have %d\n%v", i, want, have, cmp.Diff(tt.Diffs, diffs))
		}
		for k := 0; k < len(diffs); k++ {
			if want, have := tt.Diffs[k].Mode, diffs[k].Mode; want != have {
				t.Fatalf("#%d: Diffs[%d].Mode: want %v, have %v\n%v", i, k, want, have, cmp.Diff(tt.Diffs[k], diffs[k]))
			}
			if want, have := tt.Diffs[k].Src, diffs[k].Src; !cmp.Equal(want, have) {
				t.Fatalf("#%d: Diffs[%d].Src: %v", i, k, cmp.Diff(want, have))
			}
			if want, have := tt.Diffs[k].Dst, diffs[k].Dst; !cmp.Equal(want, have) {
				t.Fatalf("#%d: Diffs[%d].Dst: %v", i, k, cmp.Diff(want, have))
			}
			switch diffs[k].Mode {
			case Unchanged:
			case Created:
			case Deleted:
				if have := diffs[k].Src; have == nil {
					t.Fatalf("#%d: Diffs[%d].Src: want %v, have %v", i, k, nil, have)
				}
				if have := diffs[k].Dst; have != nil {
					t.Fatalf("#%d: Diffs[%d].Dst: want != %v, have %v", i, k, nil, have)
				}
			case Updated:
			}
		}
	}
}
