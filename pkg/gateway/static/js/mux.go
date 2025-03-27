package js

import (
	_ "embed"

	"github.com/meln5674/minimux"
)

var (
	//go:embed novnc.js
	novnc_js []byte
	//go:embed gateway.js
	gateway_js []byte

	Handler = minimux.StaticData{
		DefaultHandler: minimux.NotFound,
		PathVar:        "path",
		StaticBytes: map[string]minimux.StaticBytes{
			"novnc.js": {
				Data:        novnc_js,
				ContentType: "text/javascript",
			},
			"gateway.js": {
				Data:        gateway_js,
				ContentType: "text/javascript",
			},
		},
	}
)
