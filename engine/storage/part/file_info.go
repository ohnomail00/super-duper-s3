package part

import (
	"io"
)

type FileInfo struct {
	F    io.ReadCloser
	Size int64
}
