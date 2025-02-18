update:
	curl -L -o plugin/catalog/static/lib/htmx.js https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js
	go list -f '{{.Dir}}' -m | xargs -I {} sh -c 'cd {} && go get -u ./...' || true
	go list -f '{{.Dir}}' -m | xargs -I {} sh -c 'cd {} && go mod tidy' || true
	go work sync
generate:
	go list -f '{{.Dir}}/...' -m | xargs go generate
lint:
	go list -f '{{.Dir}}/...' -m | xargs go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v --fix -c .golangci.yaml
structalign:
	go list -f '{{.Dir}}' -m | xargs -I {} sh -c 'cd {} && go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment -fix ./...'