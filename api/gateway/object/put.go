package object

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

func (m *Handlers) Put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bucket := r.PathValue("bucket")
	object := r.PathValue("object")
	key := bucket + "/" + object

	cl := r.Header.Get("Content-Length")
	if cl == "" {
		http.Error(w, "Mandatory Content-Length", http.StatusBadRequest)
		return
	}
	fileSize, err := strconv.ParseInt(cl, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect Content-Length", http.StatusBadRequest)
		return
	}
	slog.Debug(fmt.Sprintf("Got PUT for %s, size %d byte", key, fileSize))

	plan, err := m.uploader.Do(ctx, bucket, object, r.Body, fileSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload error: %v", err), http.StatusInternalServerError)
		return
	}
	m.db.Save(key, plan)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Successfully upload"))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to write response: %v", err))
		return
	}
}
