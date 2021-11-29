package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

//go:embed build
var buildFiles embed.FS
var files, _ = fs.Sub(buildFiles, "build")

// Register registers the ui on the root path.
func Register(r *mux.Router) {
	r.Handle("/", serveFile("index.html", "text/html"))
	r.Handle("/index.html", serveFile("index.html", "text/html"))
	r.Handle("/service-worker.js", serveFile("service-worker.js", "text/javascript"))
	r.Handle("/assets-manifest.json", serveFile("asserts-manifest.json", "application/json"))
	r.Handle("/static/{type}/{resource}", http.FileServer(http.FS(files)))

	r.Handle("/favicon.ico", serveFile("favicon.ico", "image/x-icon"))
	for _, size := range []string{"16x16", "32x32", "192x192", "256x256"} {
		fileName := fmt.Sprintf("/favicon-%s.png", size)
		r.Handle(fileName, serveFile(fileName, "image/png"))
	}
}

func serveFile(name, contentType string) http.HandlerFunc {

	file, err := files.Open(name)
	if err != nil {
		log.Panic().Err(err).Msgf("could not find %s", file)
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic().Err(err).Msgf("could not read %s", file)
	}

	return func(writer http.ResponseWriter, reg *http.Request) {
		writer.Header().Set("Content-Type", contentType)
		_, _ = writer.Write(content)
	}
}
