package server

import (
	"net/http"

	"ohara/src/internal/logger"
)

const port = "3000"

func Start(handler http.Handler, log *logger.Logger) error {
	if log != nil {
		log.Info("[server] starting server url=http://localhost:%s", port)
	}
	return http.ListenAndServe(":"+port, handler)
}
