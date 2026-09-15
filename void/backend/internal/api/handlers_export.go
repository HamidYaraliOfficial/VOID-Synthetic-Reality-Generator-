package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"void/internal/export"
	"void/internal/snapshot"
)

type exportRequest struct {
	Format string `json:"format"` // "json" | "csv" | "ndjson" | "parquet_json"
}

func (a *App) handleExport(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	var req exportRequest
	_ = decodeJSON(r, &req)
	if req.Format == "" {
		req.Format = "json"
	}

	entities := eng.Entities(0)
	dir := filepath.Join(a.Cfg.DataDir, "exports")
	_ = os.MkdirAll(dir, 0o755)
	filename := fmt.Sprintf("export-%d.%s", time.Now().UnixNano(), extFor(req.Format))
	path := filepath.Join(dir, filename)

	var err error
	switch req.Format {
	case "csv":
		err = export.ToCSV(path, entities)
	case "ndjson":
		err = export.ToNDJSON(path, entities)
	case "parquet_json":
		err = export.ToParquetCompatibleJSON(path, entities)
	default:
		err = export.ToJSON(path, entities)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"path": path, "count": len(entities), "format": req.Format})
}

func extFor(format string) string {
	switch format {
	case "csv":
		return "csv"
	case "ndjson":
		return "ndjson"
	case "parquet_json":
		return "parquet.json"
	default:
		return "json"
	}
}

func (a *App) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	eng, ok := a.simOr404(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	entities := eng.Entities(0)
	snap := snapshot.New(fmt.Sprintf("snap-%d", time.Now().UnixNano()), "", id, eng.Tick(), 0, entities, nil)

	data, err := snap.Marshal()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := a.Storage.Object.Put(r.Context(), "snapshots", snap.ID+".json", data); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, snap)
}
