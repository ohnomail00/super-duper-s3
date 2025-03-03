package object

import (
	"github.com/ohnomail00/super-duper-s3/database"
	"github.com/ohnomail00/super-duper-s3/engine/gateway"
)

type Handlers struct {
	uploader   *gateway.Uploader
	downloader *gateway.Downloader
	db         database.Store
}

func New(uploader *gateway.Uploader, downloader *gateway.Downloader, db database.Store) *Handlers {
	return &Handlers{
		uploader:   uploader,
		downloader: downloader,
		db:         db,
	}
}
