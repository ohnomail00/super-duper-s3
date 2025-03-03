package middlewares

import (
	"github.com/rs/cors"
)

var Cors *cors.Cors

func SetupCors(origins []string) {
	Cors = cors.New(cors.Options{
		AllowedOrigins: origins,
	})

}
