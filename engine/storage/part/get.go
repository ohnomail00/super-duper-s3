package part

import (
	"fmt"
	"os"
	"path/filepath"
)

type Get struct {
	storagePath string
}

func NewGet(storagePath string) *Get {
	return &Get{storagePath: storagePath}
}

func (g *Get) Get(bucket, object string, part string) (*FileInfo, error) {
	// Generate the file path for the part using the template "part_{bucket}_{object}_{part}.tmp".
	filePath := filepath.Join(g.storagePath, bucket, object, fmt.Sprintf("part_%s_%s_%s.tmp", bucket, object, part))

	// Open the part file.
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	result := &FileInfo{
		F: f,
	}
	// If possible, set the Content-Length header.
	if fi, err := f.Stat(); err == nil {
		result.Size = fi.Size()
	}

	return result, nil
}
