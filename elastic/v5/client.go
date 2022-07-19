package v5

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
	elasticv5 "gopkg.in/olivere/elastic.v5"

	"github.com/olivere/esdiff/diff"
	"github.com/olivere/esdiff/elastic"
	"github.com/olivere/esdiff/elastic/config"
)

// Client implements an Elasticsearch 5.x client.
type Client struct {
	c     *elasticv5.Client
	index string
	typ   string
	size  int
}

// NewClient creates a new Client.
func NewClient(cfg *config.Config) (*Client, error) {
	var options []elasticv5.ClientOptionFunc
	if cfg != nil {
		if cfg.URL != "" {
			options = append(options, elasticv5.SetURL(cfg.URL))
		}
		if cfg.Errorlog != "" {
			f, err := os.OpenFile(cfg.Errorlog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, errors.Wrap(err, "unable to initialize error log")
			}
			l := log.New(f, "", 0)
			options = append(options, elasticv5.SetErrorLog(l))
		}
		if cfg.Tracelog != "" {
			f, err := os.OpenFile(cfg.Tracelog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, errors.Wrap(err, "unable to initialize trace log")
			}
			l := log.New(f, "", 0)
			options = append(options, elasticv5.SetTraceLog(l))
		}
		if cfg.Infolog != "" {
			f, err := os.OpenFile(cfg.Infolog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, errors.Wrap(err, "unable to initialize info log")
			}
			l := log.New(f, "", 0)
			options = append(options, elasticv5.SetInfoLog(l))
		}
		if cfg.Username != "" || cfg.Password != "" {
			options = append(options, elasticv5.SetBasicAuth(cfg.Username, cfg.Password))
		}
		options = append(options, elasticv5.SetSniff(cfg.Sniff))
	}
	cli, err := elasticv5.NewClient(options...)
	if err != nil {
		return nil, err
	}
	c := &Client{
		c:     cli,
		index: cfg.Index,
		typ:   cfg.Type,
		size:  100,
	}
	return c, nil
}

// SetBatchSize specifies the size of a single scroll operation.
func (c *Client) SetBatchSize(size int) {
	c.size = size
}

// Iterate iterates over the index.
func (c *Client) Iterate(ctx context.Context, req *elastic.IterateRequest) (<-chan *diff.Document, <-chan error) {
	docCh := make(chan *diff.Document, 1)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			close(docCh)
			close(errCh)
		}()

		// Sorting
		var sorter elasticv5.Sorter
		if req.SortField == "" {
			sorter = elasticv5.NewFieldSort("_uid").Asc()
		} else {
			field := req.SortField
			asc := true
			if field[0] == '-' {
				field = field[1:]
				asc = false
			}
			sorter = elasticv5.NewFieldSort(field).Order(asc)
		}

		svc := c.c.Scroll(c.index).Type(c.typ).Size(c.size).SortBy(sorter)
		if req.RawQuery != "" {
			q := elasticv5.NewRawStringQuery(req.RawQuery)
			svc = svc.Query(q)
		}
		if len(req.SourceFilterInclude)+len(req.SourceFilterExclude) > 0 {
			fsc := elasticv5.NewFetchSourceContext(true).
				Include(req.SourceFilterInclude...).
				Exclude(req.SourceFilterExclude...)
			svc = svc.FetchSourceContext(fsc)
		}

		for {
			res, err := svc.Do(ctx)
			if err == io.EOF {
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			if res == nil {
				errCh <- errors.New("unexpected nil document")
				return
			}
			if res.Hits == nil {
				errCh <- errors.New("unexpected nil hits")
				return
			}
			for _, hit := range res.Hits.Hits {
				doc := new(diff.Document)
				err := json.Unmarshal(*hit.Source, &doc.Source)
				if err != nil {
					errCh <- err
					return
				}
				// Replace ID field with some other field from the document?
				if req.ReplaceField != "" {
					if val, ok := doc.Source[req.ReplaceField]; ok {
						switch v := val.(type) {
						case string:
							doc.ID = v
						case int:
							doc.ID = strconv.Itoa(v)
						case int32:
							doc.ID = strconv.FormatInt(int64(v), 10)
						case int64:
							doc.ID = strconv.FormatInt(v, 10)
						case float32:
							doc.ID = strconv.Itoa(int(v))
						case float64:
							doc.ID = strconv.Itoa(int(v))
						default:
							doc.ID = val.(string)
						}
					} else {
						errCh <- errors.New("unexpected replace-with field")
						return
					}
				} else {
					doc.ID = hit.Id
				}
				docCh <- doc
			}
		}
	}()

	return docCh, errCh
}
