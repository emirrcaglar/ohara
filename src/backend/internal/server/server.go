// cert + http(s)
package server

import (
	"crypto/tls"
	"net/http"

	"ohara/src/internal/logger"

	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	Domain  string
	Port    string
	DataDir string
}

func Start(cfg Config, handler http.Handler, log *logger.Logger) error {
	if cfg.Domain == "" {
		if log != nil {
			log.Info("[server] starting local server url=http://localhost:%s", cfg.Port)
		}
		return http.ListenAndServe(":"+cfg.Port, handler)
	}

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domain),
		Cache:      autocert.DirCache(cfg.DataDir + "/certs"),
	}

	srv := &http.Server{
		Addr:    ":https",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go func() {
		if log != nil {
			log.Info("[server] starting HTTP-to-HTTPS redirect port=80")
		}
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil && log != nil {
			log.Error("[server] HTTP challenge server failed err=%v", err)
		}
	}()

	if log != nil {
		log.Info("[server] starting secure server url=https://%s", cfg.Domain)
	}
	return srv.ListenAndServeTLS("", "")
}
