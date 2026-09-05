package httpx

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

const maxBodyBytes = 4 << 20 // 4 MiB of JSON is plenty for a design payload

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write response", "error", err)
	}
}

func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// Fail renders err in the public error shape and logs the internal cause.
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	e := AsError(err)
	if e.Status >= 500 {
		slog.ErrorContext(r.Context(), "request failed",
			"code", e.Code, "error", e.Internal, "path", r.URL.Path, "request_id", RequestID(r.Context()))
	} else {
		slog.WarnContext(r.Context(), "request rejected",
			"code", e.Code, "status", e.Status, "path", r.URL.Path, "request_id", RequestID(r.Context()))
	}
	JSON(w, e.Status, map[string]any{"error": e})
}

// Decode reads a JSON body with a hard size limit and rejects unknown fields.
func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	if err := dec.Decode(dst); err != nil {
		return BadRequest("Dữ liệu gửi lên không hợp lệ.")
	}
	return nil
}
