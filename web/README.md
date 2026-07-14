# AITriage Web Dashboard

Browser-based security dashboard for AITriage. It talks to the Go backend's REST API
(`/api/*`) served by `aitriage web`, and is embedded into the Go binary at build time.

## Tech Stack

| Concern | Choice |
| :--- | :--- |
| Framework | React 19 |
| Language | TypeScript |
| Bundler | Vite 8 |
| Styling | TailwindCSS 4 (`@tailwindcss/vite`) + `@tailwindcss/typography` |
| Routing | `react-router-dom` 7 |
| State | `zustand` |
| Data fetching | `axios` |
| i18n | `i18next` / `react-i18next` (English + Russian) |
| Icons | `lucide-react` |
| Visualization | `react-force-graph-2d`, `d3-force-3d` |

## Local Development

```bash
npm install
npm run dev        # Vite dev server on http://localhost:5173
```

The dev server proxies `/api` to the backend on `http://localhost:8080`
(see `vite.config.ts`), so run `aitriage web` in a separate terminal for a
full local stack.

## Scripts

```bash
npm run dev        # Start Vite dev server with HMR
npm run build      # Type-check (tsc -b) + production build to dist/
npm run preview    # Preview the production build locally
npm run lint       # Run ESLint
npm run format     # Prettier + eslint --fix
```

## How It Ships

The production build is compiled into the Go binary rather than served separately.
From the repository root:

```bash
make sync-web      # npm build + copy web/dist → internal/server/ui/dist
make build         # Build the Go binary with embedded assets
```

`make up` / `make enterprise-up` run `sync-web` automatically before building the
Docker image. A standalone Nginx deployment is also available via `web/Dockerfile`
and `web/nginx.conf`, which proxies `/api/` to a `backend` service.

## Project Layout

```
web/src/
├── pages/         # Route-level views
├── components/    # Reusable UI components
├── ui/            # Low-level primitives
├── hooks/         # Custom React hooks
├── services/      # API clients (axios)
├── store/         # Zustand state stores
├── locales/       # i18n resources (en, ru)
├── types.ts       # Shared TypeScript types
├── i18n.ts        # i18next configuration
└── main.tsx       # App entry point
```

## Internationalization

Translations live in `src/locales/<lang>/{translation,pages,components}.json`.
The language is auto-detected in the browser with a fallback to English
(`src/i18n.ts`). To add a language, create a new locale directory and register it
in `i18n.ts`.
