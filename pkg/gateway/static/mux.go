package static

import (
	_ "embed"
	"net/http"

	"github.com/meln5674/minimux"

	"github.com/meln5674/vnc-ws-gateway/pkg/gateway/static/html"
	"github.com/meln5674/vnc-ws-gateway/pkg/gateway/static/js"
)

var Handler = minimux.InnerMuxWithPrefix("path", &minimux.Mux{
	DefaultHandler: minimux.NotFound,
	Routes: []minimux.Route{
		minimux.
			PathWithVars("html/(.+)", "path").
			WithMethods(http.MethodGet).
			IsHandledBy(html.Handler),
		minimux.
			PathWithVars("js/(.+)", "path").
			WithMethods(http.MethodGet).
			IsHandledBy(js.Handler),
	},
})
