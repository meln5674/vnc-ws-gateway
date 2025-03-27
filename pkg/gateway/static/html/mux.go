package html

import (
	_ "embed"

	"github.com/meln5674/minimux"
)

var (
	//go:embed index.html
	index_html []byte

	Handler = minimux.StaticData{
		DefaultHandler: minimux.NotFound,
		PathVar:        "path",
		StaticBytes: map[string]minimux.StaticBytes{
			"index.html": {
				Data:        index_html,
				ContentType: "text/html",
			},
		},
	}
)
