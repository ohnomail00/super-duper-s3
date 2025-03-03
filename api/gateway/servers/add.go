package servers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/ohnomail00/super-duper-s3/engine"
)

type ServerRequest struct {
	Addr string `json:"addr"`
}

func isValidURL(str string) bool {
	parsedURL, err := url.Parse(str)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	var req ServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if !isValidURL(req.Addr) {
		http.Error(w, "Invalid address", http.StatusBadRequest)
		return
	}

	h.hr.AddNode(engine.Server{Address: req.Addr}, h.cfg.VirtualReplicas)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Server added successfully"))
}
