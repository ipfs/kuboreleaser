.PHONY: kuboreleaser
kuboreleaser:
	docker build -t kuboreleaser -f Dockerfile .

.PHONY: env
env:
	./.env.sh
