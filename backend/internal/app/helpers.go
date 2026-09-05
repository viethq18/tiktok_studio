package app

import (
	"errors"
	"io"
	"net/http"

	"github.com/tks/backend/internal/auth"
)

func mustUserID(r *http.Request) string { return auth.MustUserID(r.Context()) }

// readLimited reads at most limit bytes and reports an error if there is more,
// so an oversized upload is rejected rather than silently truncated.
func readLimited(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("payload too large")
	}
	return data, nil
}
