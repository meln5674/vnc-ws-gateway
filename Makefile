all: bin/vnc-ws-gateway

GO_FILES = $(shell find ./ -name '*.go' -not -name '*_test.go' -type f)
STATIC_FILES = $(shell find ./pkg/gateway/static -type f)

package-lock.json: package.json
	npm i --package-lock-only

pkg/ui/static/js/novnc.js: package-lock.json
	npm ci
	npm run build-novnc

bin/vnc-ws-gateway: $(GO_FILES) $(STATIC_FILES)
	go build -o $@

bin/vnc-ws-gateway.static: $(GO_FILES) $(STATIC_FILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo,osusergo -ldflags '-w -extldflags "-static"' -o $@
