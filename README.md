# Design Carousel Service

**A microservice for managing and serving an image carousel, written in Go using Gin and Logrus.**

This is used on the https://dasmlab.org 'live' to feed the live content in a "pseudo realtime" manner to the site's Vue.js Component.

This service provides a live "Carousel" of media (A Queue) that can be obtained programatically via clients.  With the URLs in the Carousel, the media can be fetched and otherwise instrumented live by clients via the API.

Includes a Swagger Page API SDK page at   http://ip:port/swagger/index.html that lets you try the APIs out "live".

**[Live Site Active Link]**(https://design-carousel.svc.dasmlab.org/swagger/index.html)

---

## High Level Overview

```mermaid
flowchart TD
    subgraph Client
      A[Quasar/Vue Frontend<br>DesignCarousel.vue]
    end
    subgraph Service
      B[DesignCarousel Service]
      B1[/carousel<br>GET/POST/DELETE/]
      B2[/serve?id=xxx]
      B3[In-Memory Queue<br>Slides]
      B4[Image Store<br>optimized images]
    end

    A -- fetches list --> B1
    B1 -- returns JSON array of slides --> A
    A -- requests image (src=/serve?id=xxx) --> B2
    B2 -- streams image data --> A
    B1 -- reads/writes --> B3
    B2 -- reads --> B4
    B1 -- writes images --> B4
```
## Features

Lightweight, blazing-fast API in Golang (Gin).

In-memory queue of slide objects (for quick prototyping, easily swapped).

Add (POST), list (GET), and delete (DELETE) slides.

Serve optimized images directly via /serve?id=xxx.

Full project structure ready for extension, CI, observability.

## Endpoints
Method	Endpoint	Purpose
GET	/carousel	List all slides (JSON queue)
POST	/carousel	Add a new slide (JSON or image upload)
DELETE	/carousel/:id	Delete a slide by ID
GET	/serve?id=xxx	Serve the optimized image by slide ID

## Quick Start
git clone https://github.com/yourorg/design-carousel-service.git
cd design-carousel-service
go mod tidy
go run ./cmd/main.go

```json
{
  "id": "abc123...",
  "image_url": "/serve?id=abc123...",
  "title": "My Demo Slide",
  "source_url": "https://github.com/...",
  "created_at": "2025-06-15T20:12:34Z"
}
```

## Directory Structure
```go
design-carousel-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── api/
│   │   └── handlers.go
│   ├── storage/
│   │   └── memory.go
│   ├── imgutil/
│   │   └── resize.go
│   └── logutil/
│       └── logutil.go
├── go.mod
├── go.sum
├── PROJECT.md
└── README.md
```

## Contributing
Contact us if you want to contribute and/or fork, we would love to hear what you're using it for! :)

## License
see LICENSES/dasmlab.license for information
