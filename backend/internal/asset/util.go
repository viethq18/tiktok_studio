package asset

import (
	"bytes"
	"io"

	"github.com/google/uuid"
)

func newUUID() string          { return uuid.NewString() }
func newReader(b []byte) io.Reader { return bytes.NewReader(b) }
