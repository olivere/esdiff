package elastic

import (
	"context"

	"github.com/olivere/esdiff/diff"
)

// Client encapsulates access to an Elasticsearch cluster.
type Client interface {
	Iterate(context.Context, *IterateRequest) (<-chan *diff.Document, <-chan error)
}

// IterateRequest specifies a request for the Iterate function.
type IterateRequest struct {
	RawQuery string
}

// ClientWithBatchSize should be implemented by clients that
// support setting the batch size for scrolling.
type ClientWithBatchSize interface {
	SetBatchSize(int)
}

// ClientWithSortField should be implemented by clients that
// support setting the sort field for scrolling.
type ClientWithSortField interface {
	SetSortField(string)
}

// ClientOption specifies the signature for setting a generic
// option for a Client.
type ClientOption func(Client)

// WithBatchSize allows setting the batch size for scrolling through
// the documents (for clients that support this).
func WithBatchSize(size int) ClientOption {
	return func(client Client) {
		c, ok := client.(ClientWithBatchSize)
		if ok {
			if size > 0 {
				c.SetBatchSize(size)
			}
		}
	}
}

// WithSortField allows setting the sort field for scrolling through
// the documents (for clients that support this).
func WithSortField(sortField string) ClientOption {
	return func(client Client) {
		c, ok := client.(ClientWithSortField)
		if ok {
			if sortField != "" {
				c.SetSortField(sortField)
			}
		}
	}
}
