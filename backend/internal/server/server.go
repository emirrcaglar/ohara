// cert + http(s)
package server

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	Domain  string
	Port    string
	DataDir string
}

func Start(cfg Config, handler http.Handler) error {
	if cfg.Domain == "" {
		fmt.Printf("Starting local server on http://localhost:%s\n", cfg.Port)
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
		fmt.Printf("Starting HTTP-to-HTTPS redirect on port 80...")
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			fmt.Printf("HTTP challenge server failed: %v", err)
		}
	}()

	fmt.Printf("Starting secure server on https://%s\n", cfg.Domain)
	return srv.ListenAndServeTLS("", "")
}
