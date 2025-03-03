package object

import (
	"fmt"
	"net/http"
)

func (m *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bucket := r.PathValue("bucket")
	object := r.PathValue("object")
	key := bucket + "/" + object

	plan, exists := m.db.Get(key)
	if !exists {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}
	var totalSize int64
	for _, part := range plan.Parts {
		totalSize += part.Length
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", totalSize))
	if err := m.downloader.Do(ctx, bucket, object, plan, w); err != nil {
		http.Error(w, fmt.Sprintf("Get error: %v", err), http.StatusInternalServerError)
		return
	}
}
