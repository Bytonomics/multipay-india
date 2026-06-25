# multipay-frontend-ts (frontend checkout/picker library) — TypeScript build system
# Dual-package build: core (vanilla TS) + react (React components with styles)

.PHONY: all build lint test check format clean help typecheck build-check

all: build

typecheck:
	@echo "Type-checking TypeScript..."
	@npx tsc --noEmit

build-check: typecheck
	@echo "Build verification completed successfully"

build:
	@echo "Building TypeScript packages..."
	@npx rollup --config rollup.config.ts

lint:
	@echo "Linting TypeScript and React code..."
	@npx eslint "src/**/*.ts" "src/**/*.tsx" "tests/**/*.ts" "tests/**/*.tsx" --max-warnings 0

test:
	@echo "Running tests..."
	@npx vitest run

check: format build-check lint test
	@echo "Full pre-commit check completed successfully"

format:
	@echo "Formatting code..."
	@npx prettier --write "src/**/*.ts" "src/**/*.tsx" "tests/**/*.ts" "tests/**/*.tsx"
	@npx eslint --fix src/ tests/

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist/
	@rm -rf node_modules/.vite/
	@rm -rf coverage/

help:
	@echo "multipay-frontend-ts targets:"
	@echo "  typecheck  - Type-check TypeScript (tsc --noEmit)"
	@echo "  build-check - Verify TypeScript compilation (typecheck only)"
	@echo "  build    - Compile TypeScript packages (core + react)"
	@echo "  lint     - Run ESLint with TypeScript + React rules"
	@echo "  test     - Run Vitest test suite"
	@echo "  check    - Full pre-commit sequence (format -> build-check -> lint -> test)"
	@echo "  format   - Format code with Prettier + ESLint --fix"
	@echo "  clean    - Remove build artifacts and cache"
