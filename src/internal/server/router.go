package server

import "net/http"

func New(baseDir string) http.Handler {
	mux := http.NewServeMux()

	s := &Server{BaseDir: baseDir}
	mux.HandleFunc("GET /manga/{name}/page/{page}", s.HandleViewPage)

	mux.HandleFunc("GET /manga/{name}/info", s.HandleMangaInfo)

	return mux
}

type Server struct {
	BaseDir string
}
