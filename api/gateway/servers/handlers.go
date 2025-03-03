package servers

import (
	"github.com/ohnomail00/super-duper-s3/config"
	"github.com/ohnomail00/super-duper-s3/engine/hash"
)

type Handlers struct {
	cfg *config.Gateway
	hr  *hash.Ring
}

func New(cfg *config.Gateway, hr *hash.Ring) *Handlers {
	return &Handlers{
		hr:  hr,
		cfg: cfg,
	}

}
