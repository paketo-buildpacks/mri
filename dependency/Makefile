.PHONY: retrieve test
retrieve:
	@go run ./retrieval \
		--buildpackTomlPath "${buildpackTomlPath}" \
		--output "${output}"

test:
	@cd test; \
	./test --tarballPath $(tarballPath) --version $(version)
