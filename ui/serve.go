package ui

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
)

var box = packr.New("ui", "build/")

// Register registers the ui on the root path.
func Register(r *mux.Router) {
	r.Handle("/", serveFile("index.html", "text/html", box))
	r.Handle("/index.html", serveFile("index.html", "text/html", box))
	r.Handle("/manifest.json", serveFile("manifest.json", "application/json", box))
	r.Handle("/service-worker.js", serveFile("service-worker.js", "text/javascript", box))
	r.Handle("/assets-manifest.json", serveFile("asserts-manifest.json", "application/json", box))
	r.Handle("/static/{type}/{resource}", http.FileServer(box))

	r.Handle("/favicon.ico", serveFile("favicon.ico", "image/x-icon", box))
	for _, size := range []string{"16x16", "32x32", "192x192", "256x256"} {
		fileName := fmt.Sprintf("/favicon-%s.png", size)
		r.Handle(fileName, serveFile(fileName, "image/png", box))
	}
}

func serveFile(name, contentType string, box *packr.Box) http.HandlerFunc {
	return func(writer http.ResponseWriter, reg *http.Request) {
		writer.Header().Set("Content-Type", contentType)
		content, err := box.Find(name)
		if err != nil {
			panic(err)
		}
		_, _ = writer.Write(content)
	}
}
