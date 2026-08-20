# Linky

Linky is a link shortener that lets you create custom-named redirects using any domain you own. Instead of random short codes, you pick the alias yourself — `https://yourdomain.com/my-form` can redirect to a Google Form, a landing page, anything.

**Live demo:** [linky-shortner.vercel.app](https://linky-shortner.vercel.app)

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Backend](#backend)
  - [Frontend](#frontend)
- [Environment Variables](#environment-variables)
- [Deployment](#deployment)
- [Contribution](#contribution)

## Features

- [x] **Custom Named Links** — pick your own alias instead of a random short code.
- [x] **Link Expiry** — set an expiry date, or mark a link as permanent.
- [x] **Rate Limiting** — `/add-link` is capped per IP to stop abuse of the public write endpoint.
- [x] **CORS Allowlist** — API only accepts browser requests from configured origins.
- [ ] **QR Code Generation** — automatic QR code for each shortened link.
- [ ] **Analytics** — click tracking per link.

## Architecture

```
┌─────────────┐        ┌──────────────┐        ┌───────────┐
│   Frontend   │───────▶│    Backend    │───────▶│   Redis   │
│  React/Vite  │  HTTP  │  Go (gorilla  │        │  (links + │
│              │        │     /mux)     │        │  expiry)  │
└─────────────┘        └──────────────┘        └───────────┘
```

- **Backend** (`/backend`) — Go service that stores each short link as a Redis hash, with expiry set via `EXPIRE` when the user provides a date. Exposes `/add-link`, `/{link}` (redirect), `/links/all`, and `/health`.
- **Frontend** (`/frontend`) — React + Vite + MUI single-page app for creating links and copying the result.

## Getting Started

### Prerequisites

| Backend                          | Frontend                          |
|-----------------------------------|------------------------------------|
| Go 1.20+                          | Node.js + npm                      |
| Redis (local install or Docker)  | Familiarity with React/Vite        |
| Docker + Docker Compose (recommended) | |

### Backend

The easiest way to run the backend + Redis together locally is Docker Compose:

```bash
cd backend
docker compose up -d --build
```

This builds the Go service and starts a Redis container alongside it, both on the default Compose network. The API is available at `http://localhost:4000` — check `http://localhost:4000/health`.

To run it without Docker, start Redis yourself and then:

```bash
cd backend
go run main.go
```

See [Environment Variables](#environment-variables) below for how it connects to Redis and what CORS origins it accepts.

### Frontend

```bash
cd frontend
cp .env.example .env   # then edit VITE_API_URL if needed
npm install
npm run dev
```

This starts the Vite dev server (default `http://localhost:5173`) pointed at the backend URL set in `.env`.

## Environment Variables

**Backend** (`backend/`)

| Variable          | Default              | Purpose                                                                 |
|--------------------|----------------------|--------------------------------------------------------------------------|
| `REDIS_URL`        | —                     | Full Redis connection URL (e.g. from a managed Redis provider). Takes priority over `REDIS_ADDR` if set. |
| `REDIS_ADDR`       | `:6379`               | Host:port for Redis, used when `REDIS_URL` isn't set.                   |
| `REDIS_PASSWORD`   | empty                 | Redis auth password, used with `REDIS_ADDR`.                            |
| `ALLOWED_ORIGINS`  | `http://localhost:5173` | Comma-separated list of origins allowed to call the API from a browser. |
| `PORT`             | `4000`                | Port the server listens on (platforms like Railway inject this).        |

**Frontend** (`frontend/`)

| Variable        | Default                 | Purpose                                  |
|------------------|--------------------------|--------------------------------------------|
| `VITE_API_URL`   | `http://localhost:4000` | Base URL of the backend API. Baked in at build time. |

## Deployment

Linky's own instance runs on:

- **Backend + Redis** — [Railway](https://railway.app), deployed straight from the `backend/` Dockerfile with a Redis plugin service in the same project. `REDIS_URL` and `ALLOWED_ORIGINS` are set as service variables.
- **Frontend** — [Vercel](https://vercel.com), deployed from `frontend/` with `VITE_API_URL` set to the Railway backend's public domain.

Both are wired to auto-deploy on push to `main`. A local production-style test is also available via `backend/docker-compose.yaml` for the backend, and `npm run build && npm run preview` for the frontend.

If you're deploying your own instance, remember:
- `ALLOWED_ORIGINS` on the backend must include your deployed frontend's URL, or the browser will block requests with a CORS error.
- `VITE_API_URL` on the frontend must point at your backend's public URL — since Vite bakes it in at build time, changing it requires a rebuild, not just a restart.

## Contribution

Contributions are always welcome! To contribute:

- **Fork** the repository on GitHub.
- **Create** a new branch from `main` for your feature or bug fix:
  ```bash
  git checkout -b your-branch-name/your-name
  ```
- Check the [Issues](https://github.com/YuvanshPathak/Linky/issues) section for open issues, and comment to get one assigned to you.
- Implement your changes, then commit:
  ```bash
  git commit -m "feature XYZ implemented"
  ```
- **Push** your changes to your fork:
  ```bash
  git push origin your-branch-name/your-name
  ```
- Open a **pull request** with a description of your changes.
