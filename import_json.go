package crate

import (
	"context"
	"strings"
)

type ImportOptions struct {
	FailFast            bool
	WaitForCompletion   bool
	Shared              bool
	CompressionGzip     bool
	OverwriteDuplicates bool
}

func (db *Crate) ImportJSON(table string, url string, options ImportOptions) (err error) {
	var b strings.Builder
	var args []any
	b.Grow(100)

	b.WriteString("COPY ")
	writeIdentifier(&b, table)
	b.WriteString(" FROM ")
	writeParam(&b, &args, url)
	b.WriteString(" WITH (format='json'")

	b.WriteString(", fail_fast=")
	writeParam(&b, &args, options.FailFast)

	b.WriteString(", wait_for_completion=")
	writeParam(&b, &args, options.WaitForCompletion)

	b.WriteString(", shared=")
	writeParam(&b, &args, options.Shared)

	b.WriteString(", overwrite_duplicates=")
	writeParam(&b, &args, options.OverwriteDuplicates)

	if options.CompressionGzip {
		b.WriteString(", compression='gzip'")
	}

	b.WriteString(")")

	_, err = db.pool.Query(context.Background(), b.String(), args...)

	return
}
