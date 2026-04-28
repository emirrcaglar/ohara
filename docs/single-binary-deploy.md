# Single-Binary Deploy Walkthrough

This document explains what was changed to make Ohara deploy as a single backend binary that also serves the frontend.

## Goal

The goal was to stop treating the frontend and backend as two separate deploy artifacts.

After these changes:

- the Vue frontend is built into static files
- those static files are placed inside the backend tree
- the Go backend embeds those files into the compiled executable
- the final deploy artifact is one Go binary

## What Changed

### 1. Frontend build output can now target the backend embed directory

File changed:

- `frontend/package.json`

Added script:

```json
"build:embed": "vite build --outDir ../backend/ui/dist"
```

What it does:

- builds the Vue app with Vite
- writes the generated files into `backend/ui/dist`
- puts the frontend output where the Go backend can embed it

## 2. The backend now serves the built SPA from embedded files

Files changed:

- `backend/ui/file.go`
- `backend/internal/router/router.go`

What was added in `backend/ui/file.go`:

- a helper named `SPAHandler()`
- it reads the embedded `dist` directory using `fs.Sub`
- it serves built frontend assets with `http.FileServer`
- if the requested path does not match a real file, it falls back to `index.html`

Why the fallback matters:

- Vue Router uses client-side routes like `/library` and `/reader`
- those routes do not exist as physical files in the binary
- without an index fallback, direct navigation to those URLs would return 404

What changed in `backend/internal/router/router.go`:

- existing API and media routes were kept in place
- the SPA handler was mounted on `/`
- Go route specificity means the API routes still win over the catch-all root handler

That means:

- `/api/...` still goes to the Go handlers
- `/audio/...` still streams media
- `/library` and `/reader` now return the embedded frontend app

## 3. The frontend build was blocked by a missing asset

File added:

- `frontend/src/assets/active-transfers.svg`

Why this was needed:

- `frontend/src/views/LibraryView.vue` imported `../assets/active-transfers.svg`
- that file did not exist
- Vite could not complete a production build until the asset existed

This was not directly about embedding, but it had to be fixed for the single-binary pipeline to work.

## 4. The deploy script was updated to build the embedded frontend first

File changed:

- `deploy/deploy.example.sh`

New flow:

1. install frontend dependencies
2. build the frontend into `backend/ui/dist`
3. build the Go backend binary
4. upload and restart the service

This ensures the binary always contains the current frontend build.

## 5. Documentation was updated

File changed:

- `readme.md`

The production build section now describes the actual single-binary flow.

## Build Flow

Use these commands locally:

```bash
cd frontend
npm install
npm run build:embed

cd ../backend
go build -o ohara ./cmd
```

Result:

- `backend/ui/dist` contains the built frontend files
- `ohara` is a single binary that serves both API and frontend

## Validation Performed

The following were validated during the change:

- `npm run build:embed` succeeded
- `go build -o ./tmp/ohara.exe ./cmd` succeeded

That confirms:

- the frontend can be built into the backend tree
- the backend can compile with the embedded frontend present

## Important Notes

### Frontend dev mode is still separate

This change does not remove the normal frontend development workflow.

You can still run:

```bash
cd frontend
npm run dev
```

That is still useful for fast UI development with Vite.

### Deploy artifact is now one binary

For deployment, you no longer need a separate static hosting step for the frontend if you use the embed build flow.

You only need:

- the compiled Go binary
- the runtime data directory such as `app-data`

### The embedded frontend only updates when you rebuild it

If you change frontend code, you must rebuild the embedded frontend before rebuilding the backend binary.

In practice:

1. `npm run build:embed`
2. `go build ...`

## Files Touched

- `frontend/package.json`
- `frontend/src/assets/active-transfers.svg`
- `backend/ui/file.go`
- `backend/internal/router/router.go`
- `deploy/deploy.example.sh`
- `readme.md`

## Summary

The core idea is simple:

1. build the Vue app into `backend/ui/dist`
2. embed that directory into Go
3. serve it from the backend with SPA fallback
4. ship one backend binary

That is what turned Ohara into a single-binary deploy for frontend and backend together.