package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDataDir     = "/data"
	defaultPreloadDir  = "/app/preload_images"
	manifestFileName   = "slides.json"
	imageSubdirName    = "carousel_images"
	preloadIDNamespace = "preload:"
)

var (
	dataDir      string
	preloadDir   string
	manifestPath string
)

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func preloadSlideID(filename string) string {
	sum := sha256.Sum256([]byte(preloadIDNamespace + strings.ToLower(filename)))
	return hex.EncodeToString(sum[:16])
}

func ensureStorageDirs() error {
	if err := os.MkdirAll(imageBasePath, 0o775); err != nil {
		return err
	}
	return os.MkdirAll(dataDir, 0o775)
}

func loadManifest() error {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var slides []Slide
	if err := json.Unmarshal(raw, &slides); err != nil {
		return err
	}

	storeMu.Lock()
	defer storeMu.Unlock()
	for _, slide := range slides {
		if slide.ID == "" {
			continue
		}
		slideStore[slide.ID] = slide
	}
	return nil
}

func persistManifest() error {
	storeMu.RLock()
	slides := make([]Slide, 0, len(slideStore))
	for _, slide := range slideStore {
		slides = append(slides, slide)
	}
	storeMu.RUnlock()

	payload, err := json.MarshalIndent(slides, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o664); err != nil {
		return err
	}
	return os.Rename(tmpPath, manifestPath)
}

func removeSlideImage(id string) {
	imagePath := filepath.Join(imageBasePath, id+".webp")
	_ = os.Remove(imagePath)
}

// Initialize prepares writable storage, restores persisted slides, and seeds preload images.
func Initialize() error {
	dataDir = envOrDefault("CAROUSEL_DATA_DIR", defaultDataDir)
	preloadDir = envOrDefault("CAROUSEL_PRELOAD_DIR", defaultPreloadDir)
	imageBasePath = filepath.Join(dataDir, imageSubdirName)
	manifestPath = filepath.Join(dataDir, manifestFileName)

	if err := ensureStorageDirs(); err != nil {
		return err
	}

	if err := loadManifest(); err != nil {
		log.Warnf("Storage: could not load manifest %s: %v", manifestPath, err)
	}

	before := slideCount()
	seeded, err := PreloadImagesFromDir(preloadDir)
	if err != nil {
		log.Warnf("Storage: preload from %s failed: %v", preloadDir, err)
	}

	after := slideCount()
	if seeded > 0 || (before == 0 && after > 0) {
		if err := persistManifest(); err != nil {
			log.Warnf("Storage: could not persist manifest: %v", err)
		}
	}

	log.Infof(
		"Storage: ready data_dir=%s images=%s manifest=%s slides=%d preloaded_now=%d",
		dataDir,
		imageBasePath,
		manifestPath,
		after,
		seeded,
	)
	return nil
}

func slideCount() int {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return len(slideStore)
}
