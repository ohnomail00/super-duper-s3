package part

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) != 3 {
		http.Error(w, "Invalid path, expected /{bucket}/{object}/{part}", http.StatusBadRequest)
		return
	}
	bucket, object, part := parts[0], parts[1], parts[2]

	file, err := h.parts.get.Get(bucket, object, part)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		if cerr := file.F.Close(); cerr != nil {
			slog.Error("failed to close file", "error", cerr)
		}
	}()

	if _, err := io.Copy(w, file.F); err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}
}
