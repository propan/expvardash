fmt:
	find . -type f -name '*.go' -not -path "./vendor/*" -exec goimports -w -l {} +
