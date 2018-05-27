PKGS = $(shell go list ./... | grep -v /vendor/)

fmt:
	find . -type f -name '*.go' -not -path "./vendor/*" -exec goimports -w -l {} +

godep:
	godep save $(PKGS)

bundle:
	go-bindata -o resources.go templates static/...

test:
	go test -race -cover $(PKGS)

install:
	go install $(PKGS)
