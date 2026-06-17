package api

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
)

func isSupportedPreloadImage(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg")
}

// PreloadImagesFromDir imports baked-in images into persistent storage.
// Existing slide IDs (stable per filename) are skipped so restarts are idempotent.
func PreloadImagesFromDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	seeded := 0
	for _, entry := range entries {
		if entry.IsDir() || !isSupportedPreloadImage(entry.Name()) {
			continue
		}

		id := preloadSlideID(entry.Name())
		storeMu.RLock()
		_, exists := slideStore[id]
		storeMu.RUnlock()
		if exists {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		imgData, err := os.ReadFile(fullPath)
		if err != nil {
			log.Warnf("Preload: failed to read %s: %v", entry.Name(), err)
			continue
		}

		webpFilename := filepath.Join(imageBasePath, id+".webp")
		img, _, err := image.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Warnf("Preload: could not decode %s: %v", entry.Name(), err)
			continue
		}

		outFile, err := os.Create(webpFilename)
		if err != nil {
			log.Warnf("Preload: failed to write %s: %v", webpFilename, err)
			continue
		}

		options := &webp.Options{Lossless: false, Quality: 82}
		if err := webp.Encode(outFile, img, options); err != nil {
			outFile.Close()
			log.Warnf("Preload: WebP encode failed for %s: %v", entry.Name(), err)
			continue
		}
		outFile.Close()

		slide := Slide{
			ID:        id,
			Title:     "Preloaded: " + entry.Name(),
			SourceURL: "",
			ImageURL:  "/serve?id=" + id,
			CreatedAt: time.Now().UTC(),
		}

		storeMu.Lock()
		slideStore[id] = slide
		storeMu.Unlock()
		log.Infof("Preload: added %s as slide ID=%s", entry.Name(), id)
		seeded++
	}

	log.Infof("Preload: seeded %d image(s) from %s", seeded, dir)
	return seeded, nil
}
