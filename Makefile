.PHONY: dev stop install

dev:
	@echo "🚀 Starting backend & frontend..."
	@(cd server && task dev) &
	@(cd web && pnpm dev) &
	@wait

stop:
	@echo "🛑 Stopping dev servers..."
	@pkill -f "air" || true
	@pkill -f "vite" || true
	@echo "✓ Stopped"

install:
	@(cd server && go mod tidy)
	@(cd web && pnpm install)
