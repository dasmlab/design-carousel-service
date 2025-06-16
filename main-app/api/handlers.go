// internal/api/handlers.go

package api

import (
	// STD LIBS
	"os"
	"io"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"bytes"
	"path/filepath"
	"net/http"
	"sync"
	"time"
	"strings"

	// 3PP's 
	"github.com/chai2010/webp" 
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	// Our Stuff
	"design-carousel-service/logutil"
)

// Slide represents a carousel entry
type Slide struct {
	ID        string    `json:"id" example:"abc123"`
	ImageURL  string    `json:"image_url" example:"/serve?id=abc123"`
	Title     string    `json:"title" example:"Swiss Alps"`
	SourceURL string    `json:"source_url" example:"https://en.wikipedia.org/wiki/Fronalpstock"`
	CreatedAt time.Time `json:"created_at" example:"2025-06-15T20:12:34Z"`
}

// Thread-safe in-memory slide store
var (
	slideStore = make(map[string]Slide)
	storeMu    sync.RWMutex
	imageBasePath = "./carousel_images"
	componentName = "design-carousel-api"
	log = logutil.InitLogger(componentName)
)

func init() {
	os.MkdirAll(imageBasePath, 0755)
	preloadDir := "./preload_images"
	PreloadImagesFromDir(preloadDir)
}

//
// 1. List all slides
//

// ListSlides godoc
// @Summary      List all slides in the carousel
// @Description  Returns the full queue of slide entries
// @Tags         carousel
// @Produce      json
// @Success      200 {array} Slide
// @Router       /carousel [get]
func ListSlides(c *gin.Context) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	slides := make([]Slide, 0, len(slideStore))
	for _, s := range slideStore {
		slides = append(slides, s)
	}
	c.JSON(http.StatusOK, slides)
}

// AddSlide supports both JSON and multipart file upload

// @Summary      Add a new slide (JSON or file upload)
// @Description  Adds a slide entry. Accepts multipart form (file+meta) or JSON.
// @Tags         carousel
// @Accept       json,mpfd
// @Produce      json
// @Param        title      formData  string  false  "Slide Title"
// @Param        source_url formData  string  false  "Source URL"
// @Param        image      formData  file    false  "Image file (png/jpeg)"
// @Success      201 {object} Slide
// @Failure      400 {object} map[string]string
// @Router       /carousel [post]
func AddSlide(c *gin.Context) {
    contentType := c.ContentType()
    var title, sourceURL string
    var imgData []byte
    var fileName, imgExt string

    log.Infof("AddSlide: Received content-type: %s", contentType)

    // Multipart/file upload support
    if strings.HasPrefix(contentType, "multipart/") {
        file, header, err := c.Request.FormFile("image")
        if err == nil && file != nil {
            defer file.Close()
            imgData, _ = io.ReadAll(file)
            fileName = header.Filename
            imgExt = filepath.Ext(fileName)
            log.Infof("AddSlide: Uploaded file: %s (%d bytes), ext=%s", fileName, len(imgData), imgExt)
        } else {
            log.Warnf("AddSlide: No file uploaded or error reading file: %v", err)
        }
        title = c.PostForm("title")
        sourceURL = c.PostForm("source_url")
    } else {
        // JSON fallback
        var req struct {
            Title     string `json:"title"`
            SourceURL string `json:"source_url"`
            ImageURL  string `json:"image_url"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            log.Warnf("AddSlide: Bad JSON or missing fields: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "No valid image or JSON: " + err.Error()})
            return
        }
        title = req.Title
        sourceURL = req.SourceURL
        log.Infof("AddSlide: JSON only request, no image uploaded (title=%s)", title)
        // Optional: fetch from ImageURL if provided (future feature)
    }

    if imgData == nil {
        log.Warnf("AddSlide: No image data found in request")
        c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded or referenced"})
        return
    }

    // Decode image, resize, encode as WebP
    id := uuid.NewString()
    webpFilename := filepath.Join(imageBasePath, id+".webp")

    img, format, err := image.Decode(bytes.NewReader(imgData))
    if err != nil {
        log.Errorf("AddSlide: Image decode failed: %v (filename: %s)", err, fileName)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Image decode failed: " + err.Error()})
        return
    }
    bounds := img.Bounds()
    log.Infof("AddSlide: Image decoded: format=%s, dimensions=%dx%d", format, bounds.Dx(), bounds.Dy())

    outFile, err := os.Create(webpFilename)
    if err != nil {
        log.Errorf("AddSlide: Failed to store image at %s: %v", webpFilename, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store image"})
        return
    }
    defer outFile.Close()

    options := &webp.Options{Lossless: false, Quality: 82}
    if err := webp.Encode(outFile, img, options); err != nil {
        log.Errorf("AddSlide: WebP encode failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "WebP encode failed"})
        return
    }
    log.Infof("AddSlide: Image written to %s (id=%s)", webpFilename, id)

    slide := Slide{
        ID:        id,
        Title:     title,
        SourceURL: sourceURL,
        ImageURL:  "/serve?id=" + id,
        CreatedAt: time.Now().UTC(),
    }

    storeMu.Lock()
    slideStore[id] = slide
    storeMu.Unlock()

    log.Infof("AddSlide: Slide added (ID=%s, title=%s, url=%s)", id, title, slide.ImageURL)
    c.JSON(http.StatusCreated, slide)
}

//
// 3. Delete slide by ID
//

// DeleteSlide godoc
// @Summary      Delete a slide
// @Description  Removes a slide from the carousel by its unique ID
// @Tags         carousel
// @Produce      json
// @Param        id   path      string  true  "Slide ID"
// @Success      204  "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /carousel/{id} [delete]
func DeleteSlide(c *gin.Context) {
	id := c.Param("id")
	storeMu.Lock()
	defer storeMu.Unlock()

	if _, ok := slideStore[id]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Slide not found"})
		return
	}
	delete(slideStore, id)
	c.Status(http.StatusNoContent)
}

// ServeImage streams the optimized image
// @Summary      Serve the optimized image by slide ID
// @Description  Streams the optimized (WebP) image for the given slide ID
// @Tags         serve
// @Produce      image/webp
// @Param        id query string true "Slide ID"
// @Success      200 {file} binary
// @Failure      404 {object} map[string]string
// @Router       /serve [get]
func ServeImage(c *gin.Context) {
	id := c.Query("id")

	storeMu.RLock()
	_, ok := slideStore[id]
	storeMu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	imagePath := filepath.Join(imageBasePath, id+".webp")
	f, err := os.Open(imagePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image file not found"})
		return
	}
	defer f.Close()

	c.Header("Content-Type", "image/webp")
	c.File(imagePath)
}

