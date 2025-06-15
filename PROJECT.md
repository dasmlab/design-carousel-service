# DesignCarousel Service – Project Overview

## Purpose

DesignCarousel is a modern, lightweight Go (Golang) microservice for managing, serving, and curating a horizontal image carousel for the DASMLAB UI and other dashboards. It provides APIs for uploading, optimizing, listing, and deleting images (slides), as well as serving optimized images by ID.

This project also serves as a reference template for building future DASMLAB microservices with:
- Structured project layout (`cmd/`, `internal/`, `logutil/`, etc.)
- Gin-based REST API
- Logrus for structured logging
- In-memory (swappable) storage layer
- Ready-to-extend for Prometheus, OpenAPI, image processing, SSE, etc.

## High-Level API Design

- `GET    /carousel`  
  Returns the full list of carousel slides (queue), as JSON.

- `POST   /carousel`  
  Adds a new slide. Accepts either a JSON payload with an image URL, or a file upload (multipart). Images are resized/compressed to optimal web format (WebP or JPEG).

- `DELETE /carousel/:id`  
  Removes a slide by its unique ID.

- `GET    /serve?id=XXX`  
  Serves (streams) the optimized image (shrunken for bandwidth) by slide ID.  
  (Used as the image `src` in the client carousel UI.)

## Storage

- Current implementation uses in-memory storage (see `internal/storage/memory.go`).
- Production: can be swapped for S3, Minio, or any blob backend.

## Intended Extension Points

- Add persistent backend (BoltDB, Postgres, S3, etc.)
- Extend `/carousel` POST to accept file uploads and run auto-resize to WebP.
- Implement authentication/authorization for slide modification.
- Add Prometheus metrics middleware, tracing, etc.
- SSE support for UI “live update” (hot reload on change).
- Generate OpenAPI/Swagger spec for auto-doc.

## Project Directory Structure

design-carousel-service/
├── cmd/
│ └── main.go
├── internal/
│ ├── api/
│ │ └── handlers.go
│ ├── storage/
│ │ └── memory.go
│ ├── imgutil/
│ │ └── resize.go # (optional, image manipulation utils)
│ └── logutil/
│ └── logutil.go
├── go.mod
├── go.sum
├── PROJECT.md
└── README.md

yaml
Copy
Edit

---

## Example Slide JSON

```json
{
  "id": "fa6f1e...",
  "image_url": "/serve?id=fa6f1e...",    // (served from local service)
  "title": "Swiss Alps",
  "source_url": "https://en.wikipedia.org/wiki/Fronalpstock",
  "created_at": "2025-06-15T20:12:34Z"
}
