package part

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Put struct {
	storagePath string
}

func NewPut(storagePath string) *Put {
	return &Put{storagePath: storagePath}
}

func (p *Put) Put(bucket, object, part string, content io.Reader) error {
	// Create directory for bucket if it doesn't exist.
	if err := os.MkdirAll(filepath.Join(p.storagePath, bucket, object), 0755); err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	// Generate the path for saving the part.
	destPath := filepath.Join(p.storagePath, bucket, object, fmt.Sprintf("part_%s_%s_%s.tmp", bucket, object, part))

	// Create (or overwrite) the part file.
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil {
			slog.Error("failed to close file", "error", cerr)
		}
	}()

	// Copy request content into the file.
	if _, err := io.Copy(f, content); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
