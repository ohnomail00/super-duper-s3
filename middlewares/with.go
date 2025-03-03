package middlewares

import (
	"net/http"
)

func With(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return TraceMiddleware(
		LogMiddleware(
			Cors.Handler(
				http.HandlerFunc(handler))))
}
