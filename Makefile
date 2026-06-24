# multipay-india — top-level dispatcher Makefile.
# Runs the same target in every language subdirectory that has a Makefile.
# Placeholder ports (multipay-ts, multipay-py, multipay-frontend-ts) ship stub
# Makefiles, so every target runs cleanly across the whole monorepo.

LANGS := multipay-go multipay-ts multipay-py multipay-frontend-ts

.PHONY: all build lint test check format clean help

all: build

build:
	@for d in $(LANGS); do if [ -f $$d/Makefile ]; then echo "==> $$d: build"; $(MAKE) -C $$d build || exit 1; fi; done

lint:
	@for d in $(LANGS); do if [ -f $$d/Makefile ]; then echo "==> $$d: lint"; $(MAKE) -C $$d lint || exit 1; fi; done

test:
	@for d in $(LANGS); do if [ -f $$d/Makefile ]; then echo "==> $$d: test"; $(MAKE) -C $$d test || exit 1; fi; done

check:
	@for d in $(LANGS); do if [ -f $$d/Makefile ]; then echo "==> $$d: check"; $(MAKE) -C $$d check || exit 1; fi; done

format:
	@for d in $(LANGS); do if [ -f $$d/Makefile ]; then echo "==> $$d: format"; $(MAKE) -C $$d format || exit 1; fi; done

clean:
	@for d in $(LANGS); do if [ -f $$d/Makefile ]; then echo "==> $$d: clean"; $(MAKE) -C $$d clean || exit 1; fi; done

help:
	@echo "multipay-india monorepo dispatcher — runs each target across language folders:"
	@echo "  $(LANGS)"
	@echo ""
	@echo "Targets: build (default), lint, test, check, format, clean, help"
