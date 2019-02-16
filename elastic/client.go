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
	RawQuery            string
	SortField           string
	SourceFilterInclude []string
	SourceFilterExclude []string
}

// ClientWithBatchSize should be implemented by clients that
// support setting the batch size for scrolling.
type ClientWithBatchSize interface {
	SetBatchSize(int)
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
