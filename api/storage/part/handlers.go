package part

import (
	"strings"

	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/engine/storage/part"
)

type parts struct {
	get *part.Get
	put *part.Put
}

type Handlers struct {
	cfg   *config.Storage
	parts parts
}

func NewHandlers(cfg *config.Storage, get *part.Get, put *part.Put) *Handlers {
	return &Handlers{
		cfg: cfg,
		parts: parts{
			get: get,
			put: put,
		},
	}

}

func splitPath(p string) []string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}
