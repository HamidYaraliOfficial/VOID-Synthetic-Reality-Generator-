// Package export writes simulation results / synthetic datasets to disk in
// formats suited to downstream Big Data processing: JSON, CSV, NDJSON and a
// columnar format compatible with Parquet-style analytical tooling.
package export

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"void/internal/entity"
)

// ToJSON writes entities as a single JSON array.
func ToJSON(path string, entities []entity.Entity) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(entities)
}

// ToNDJSON writes one JSON object per line - the preferred format for
// streaming into Big Data pipelines (Spark, BigQuery, etc.).
func ToNDJSON(path string, entities []entity.Entity) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	enc := json.NewEncoder(w)
	for _, e := range entities {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

// ToCSV flattens each entity's Attrs map into columns (a union of all attr
// keys seen across the population, sorted for stable column ordering) plus
// standard identity columns.
func ToCSV(path string, entities []entity.Entity) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	attrKeys := collectAttrKeys(entities)
	w := csv.NewWriter(f)
	defer w.Flush()

	header := append([]string{"id", "archetype", "created_tick", "status"}, attrKeys...)
	if err := w.Write(header); err != nil {
		return err
	}
	for _, e := range entities {
		row := []string{e.ID, e.Archetype, fmt.Sprintf("%d", e.CreatedAt), e.State["status"]}
		for _, k := range attrKeys {
			row = append(row, fmt.Sprintf("%v", e.Attrs[k]))
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func collectAttrKeys(entities []entity.Entity) []string {
	set := map[string]bool{}
	for _, e := range entities {
		for k := range e.Attrs {
			set[k] = true
		}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ColumnarBatch is a simplified columnar representation compatible with
// Parquet-style analytical consumption: one slice per attribute column.
// Bundled without an external Parquet dependency; production pipelines can
// feed ColumnarBatch straight into a parquet-go writer behind this same
// shape without touching simulation/export call sites.
type ColumnarBatch struct {
	IDs     []string             `json:"ids"`
	Columns map[string][]float64 `json:"columns"`
}

func ToColumnar(entities []entity.Entity) ColumnarBatch {
	keys := collectAttrKeys(entities)
	cb := ColumnarBatch{IDs: make([]string, len(entities)), Columns: map[string][]float64{}}
	for _, k := range keys {
		cb.Columns[k] = make([]float64, len(entities))
	}
	for i, e := range entities {
		cb.IDs[i] = e.ID
		for _, k := range keys {
			cb.Columns[k][i] = e.Attrs[k]
		}
	}
	return cb
}

// ToParquetCompatibleJSON writes the ColumnarBatch as JSON - a portable,
// dependency-free stand-in that any real Parquet converter can ingest.
func ToParquetCompatibleJSON(path string, entities []entity.Entity) error {
	cb := ToColumnar(entities)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cb)
}
