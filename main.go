package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/olivere/esdiff/diff"
	"github.com/olivere/esdiff/diff/printer"
	"github.com/olivere/esdiff/elastic"
	"github.com/olivere/esdiff/elastic/config"
	v5 "github.com/olivere/esdiff/elastic/v5"
	v6 "github.com/olivere/esdiff/elastic/v6"
	v7 "github.com/olivere/esdiff/elastic/v7"
)

func main() {
	var (
		outputFormat     = flag.String("o", "", "Output format, e.g. json")
		size             = flag.Int("size", 100, "Batch size")
		rawSrcQuery      = flag.String("sf", "", `Raw query for filtering the source, e.g. {"term":{"user":"olivere"}}`)
		rawDstQuery      = flag.String("df", "", `Raw query for filtering the destination, e.g. {"term":{"name.keyword":"Oliver"}}`)
		srcSort          = flag.String("ssort", "", `Field to sort the source, e.g. "id" or "-id" (prepend with - for descending)`)
		dstSort          = flag.String("dsort", "", `Field to sort the destination, e.g. "id" or "-id" (prepend with - for descending)`)
		srcFilterInclude = flag.String("include", "", `Raw source filter for including certain fields from the source, e.g. "obj.*"`)
		srcFilterExclude = flag.String("exclude", "", `Raw source filter for excluding certain fields from the source, e.g. "hash_value,sub.*"`)
		unchanged        = flag.Bool("u", false, `Print unchanged docs`)
		updated          = flag.Bool("c", true, `Print changed docs`)
		changed          = flag.Bool("a", true, `Print added docs`)
		deleted          = flag.Bool("d", true, `Print deleted docs`)
	)

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(1)
	}

	var srcFilterIncludes []string
	if *srcFilterInclude != "" {
		srcFilterIncludes = strings.Split(*srcFilterInclude, ",")
	}
	var srcFilterExcludes []string
	if *srcFilterExclude != "" {
		srcFilterExcludes = strings.Split(*srcFilterExclude, ",")
	}

	options := []elastic.ClientOption{
		elastic.WithBatchSize(*size),
	}

	src, err := newClient(flag.Arg(0), options...)
	if err != nil {
		log.Fatal(err)
	}
	srcIterReq := &elastic.IterateRequest{
		RawQuery:            *rawSrcQuery,
		SortField:           *srcSort,
		SourceFilterInclude: srcFilterIncludes,
		SourceFilterExclude: srcFilterExcludes,
	}

	dst, err := newClient(flag.Arg(1), options...)
	if err != nil {
		log.Fatal(err)
	}
	dstIterReq := &elastic.IterateRequest{
		RawQuery:            *rawDstQuery,
		SortField:           *dstSort,
		SourceFilterInclude: srcFilterIncludes,
		SourceFilterExclude: srcFilterExcludes,
	}

	var p printer.Printer
	{
		switch *outputFormat {
		default:
			p = printer.NewStdPrinter(os.Stdout, *unchanged, *updated, *changed, *deleted)
		case "json":
			p = printer.NewJSONPrinter(os.Stdout, *unchanged, *updated, *changed, *deleted)
		}
	}

	g, ctx := errgroup.WithContext(context.Background())
	srcDocCh, srcErrCh := src.Iterate(ctx, srcIterReq)
	_ = srcErrCh
	dstDocCh, dstErrCh := dst.Iterate(ctx, dstIterReq)
	_ = dstErrCh
	diffCh, errCh := diff.Differ(ctx, srcDocCh, dstDocCh)
	_ = errCh
	g.Go(func() error {
		for {
			select {
			case d, ok := <-diffCh:
				if !ok {
					return nil
				}
				if err := p.Print(d); err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	g.Go(func() error {
		return <-srcErrCh
	})
	g.Go(func() error {
		return <-dstErrCh
	})
	g.Go(func() error {
		return <-errCh
	})
	if err = g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "General usage:\n\n")
	fmt.Fprintf(os.Stderr, "\t%s [flags] <source-url> <destination-url>\n\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "General flags:\n")
	flag.PrintDefaults()
}

// newClient will create a new Elasticsearch client,
// matching the supported version.
func newClient(url string, opts ...elastic.ClientOption) (elastic.Client, error) {
	cfg, err := config.Parse(url)
	if err != nil {
		return nil, err
	}
	v, major, _, _, err := elasticsearchVersion(cfg)
	if err != nil {
		return nil, err
	}
	switch major {
	case 5:
		c, err := v5.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		for _, opt := range opts {
			opt(c)
		}
		return c, nil
	case 6:
		c, err := v6.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		for _, opt := range opts {
			opt(c)
		}
		return c, nil
	case 7:
		c, err := v7.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		for _, opt := range opts {
			opt(c)
		}
		return c, nil
	default:
		return nil, errors.Errorf("unsupported Elasticsearch version %s", v)
	}
}

// elasticsearchVersion determines the Elasticsearch option.
func elasticsearchVersion(cfg *config.Config) (string, int64, int64, int64, error) {
	type infoType struct {
		Name    string `json:"name"`
		Version struct {
			Number string `json:"number"` // e.g. "6.2.4"
		} `json:"version"`
	}
	req, err := http.NewRequest("GET", cfg.URL, nil)
	if err != nil {
		return "", 0, 0, 0, err
	}
	if cfg.Username != "" || cfg.Password != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, 0, 0, err
	}
	defer res.Body.Close()
	var info infoType
	if err = json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", 0, 0, 0, err
	}
	v, err := semver.NewVersion(info.Version.Number)
	if err != nil {
		return info.Version.Number, 0, 0, 0, err
	}
	return info.Version.Number, v.Major(), v.Minor(), v.Patch(), nil
}
