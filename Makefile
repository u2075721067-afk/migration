# Context Utilities Makefile
# Provides convenient targets for context management operations

.PHONY: ctx-init ctx-save ctx-last ctx-show help

# Default target
help:
	@echo "Context Utilities - Available targets:"
	@echo "  ctx-init    - Initialize/update context schema"
	@echo "  ctx-save    - Save context with title, text, and tags"
	@echo "  ctx-last    - List recent context items (default: 10)"
	@echo "  ctx-show    - Show specific context item by ID"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make ctx-save TITLE=\"My Title\" TEXT=\"My content\" TAGS=\"demo,test\""
	@echo "  make ctx-last LIMIT=5"
	@echo "  make ctx-show ID=1"

# Initialize/update context schema
ctx-init:
	cd MOVA_EVAL && python .tools/ctx.py init

# Save context with title, text, and tags
ctx-save:
	@if [ -z "$(TITLE)" ] || [ -z "$(TEXT)" ]; then \
		echo "Error: TITLE and TEXT are required"; \
		echo "Usage: make ctx-save TITLE=\"Title\" TEXT=\"Content\" TAGS=\"tag1,tag2\""; \
		exit 1; \
	fi
	cd MOVA_EVAL && python .tools/ctx.py save --title "$(TITLE)" --text "$(TEXT)" --tags "$(TAGS)"

# List recent context items
ctx-last:
	cd MOVA_EVAL && python .tools/ctx.py last --limit $(or $(LIMIT),10)

# Show specific context item by ID
ctx-show:
	@if [ -z "$(ID)" ]; then \
		echo "Error: ID is required"; \
		echo "Usage: make ctx-show ID=1"; \
		exit 1; \
	fi
	cd MOVA_EVAL && python .tools/ctx.py show --id $(ID) --format text
