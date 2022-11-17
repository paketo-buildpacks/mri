.PHONY: test

retrieve:
	@cd retrieval; \
	go run main.go \
		--buildpackTomlPath "${buildpackTomlPath}" \
		--output "${output}"

test:
	@cd test; \
	./test --tarballPath $(tarballPath) --version $(version)
