.PHONY: run compare run-all compare-all clean help

LANG ?= go
CONFIGS_DIR := configs
RESULTS_DIR := results
SCRIPTS_DIR := scripts

## run: Run benchmark for a single language (LANG=go)
run:
	@echo "Running benchmark for $(LANG)..."
	$(SCRIPTS_DIR)/run-benchmark.sh $(CONFIGS_DIR)/$(LANG).yaml $(RESULTS_DIR)

## compare: Compare results to baseline for a single language (LANG=go)
compare:
	@echo "Comparing $(LANG) results to baseline..."
	$(SCRIPTS_DIR)/report.sh $(RESULTS_DIR)/$(LANG).json $(RESULTS_DIR)/baselines/$(LANG).json

## run-all: Run benchmarks for all configured languages
run-all:
	@for cfg in $(CONFIGS_DIR)/*.yaml; do \
		lang=$$(basename "$$cfg" .yaml); \
		echo "--- Running benchmark: $$lang ---"; \
		$(MAKE) run LANG=$$lang || exit 1; \
	done

## compare-all: Compare all results to baselines
compare-all:
	@for result in $(RESULTS_DIR)/*.json; do \
		lang=$$(basename "$$result" .json); \
		if [ -f "$(RESULTS_DIR)/baselines/$$lang.json" ]; then \
			echo "--- Comparing: $$lang ---"; \
			$(MAKE) compare LANG=$$lang || exit 1; \
		fi; \
	done

## clean: Remove cloned exemplar workdirs and latest results
clean:
	rm -rf workdir/
	rm -f $(RESULTS_DIR)/*.json

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
