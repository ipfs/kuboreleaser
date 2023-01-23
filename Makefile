.PHONY: kuboreleaser
kuboreleaser:
	docker build -t kuboreleaser -f Dockerfile .
