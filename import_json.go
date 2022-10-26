package crate

import (
	"context"
	"strings"
)

type ImportOptions struct {
	FailFast            bool `json:"failFast"`
	WaitForCompletion   bool `json:"waitForCompletion"`
	Shared              bool `json:"shared"`
	CompressionGzip     bool `json:"compressionGzip"`
	OverwriteDuplicates bool `json:"overwriteDuplicates"`
	Partition           Condition
}

type ImportResult struct {
	Node         ImportNode             `json:"node"`
	Uri          string                 `json:"uri"`
	ErrorCount   int                    `json:"error_count"`
	SuccessCount int                    `json:"success_count"`
	Errors       map[string]ImportError `json:"errors"`
}

type ImportNode struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ImportError struct {
	Count       int   `json:"count"`
	LineNumbers []int `json:"line_numbers"`
}

func (db *Crate) ImportJSON(ctx context.Context, table string, url string, options ImportOptions) (results []ImportResult, err error) {
	var b strings.Builder
	var args []any
	b.Grow(100)

	b.WriteString("COPY ")
	writeIdentifier(&b, table)

	if options.Partition != nil {
		b.WriteString(" PARTITION (")
		options.Partition.run(&b, &args)
		b.WriteString(")")
	}

	b.WriteString(" FROM ")
	writeParam(&b, &args, url)
	b.WriteString(" WITH (format='json'")

	writeBoolOption(&b, "fail_fast", options.FailFast)
	writeBoolOption(&b, "wait_for_completion", options.WaitForCompletion)
	writeBoolOption(&b, "shared", options.Shared)
	writeBoolOption(&b, "overwrite_duplicates", options.OverwriteDuplicates)

	if options.CompressionGzip {
		b.WriteString(", compression='gzip'")
	}

	b.WriteString(") RETURN SUMMARY")

	rows, err := db.pool.Query(ctx, b.String(), args...)

	if err != nil {
		return
	}

	defer rows.Close()

	results = make([]ImportResult, 0, 1)

	for rows.Next() {
		result := ImportResult{}
		err = rows.Scan(&result.Node, &result.Uri, &result.SuccessCount, &result.ErrorCount, &result.Errors)
		results = append(results, result)
	}

	return
}

func writeBoolOption(b *strings.Builder, key string, value bool) {
	b.WriteString(", ")
	b.WriteString(key)
	b.WriteByte('=')

	if value {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
}
