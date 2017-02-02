PKGS = $(shell go list ./... | grep -v /vendor/)

godep:
	godep save $(PKGS)

fmt:
	find . -type f -name '*.go' -not -path "./vendor/*" -exec goimports -w -l {} +
