package part

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (h *Handlers) Put(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) != 3 {
		http.Error(w, "Invalid path, expected /{bucket}/{object}/{part}", http.StatusBadRequest)
		return
	}
	bucket, object, part := parts[0], parts[1], parts[2]

	err := h.parts.put.Put(bucket, object, part, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to put file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Upload part successful"))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write response: %v", err))
		return
	}
}
