package api

import (
    "io/ioutil"
    "strings"
   "os"
   "image"
   _ "image/jpeg"
   _ "image/png"
   "bytes"
   "path/filepath"
   "time"

    // 3PP's
   "github.com/chai2010/webp"
   "github.com/google/uuid"

   // Our Stuff
   "design-carousel-service/logutil"

)



// Call this ONCE on startup from main.go (after os.MkdirAll(imageBasePath, 0755))
func PreloadImagesFromDir(preloadDir string) {
    log := logutil.InitLogger("design-carousel-preload")
    files, err := ioutil.ReadDir(preloadDir)
    if err != nil {
        log.Warnf("Preload: Could not read preload dir %s: %v", preloadDir, err)
        return
    }
    count := 0
    for _, f := range files {
        if f.IsDir() {
            continue
        }
        name := f.Name()
        // Only support .png/.jpg/.jpeg for now
        if !(strings.HasSuffix(strings.ToLower(name), ".png") ||
            strings.HasSuffix(strings.ToLower(name), ".jpg") ||
            strings.HasSuffix(strings.ToLower(name), ".jpeg")) {
            continue
        }
        fullPath := filepath.Join(preloadDir, name)
        imgData, err := os.ReadFile(fullPath)
        if err != nil {
            log.Warnf("Preload: Failed to read image %s: %v", name, err)
            continue
        }
        id := uuid.NewString()
        webpFilename := filepath.Join(imageBasePath, id+".webp")

        //img, format, err := image.Decode(bytes.NewReader(imgData))
        img, _, err := image.Decode(bytes.NewReader(imgData))
        if err != nil {
            log.Warnf("Preload: Could not decode %s: %v", name, err)
            continue
        }
        outFile, err := os.Create(webpFilename)
        if err != nil {
            log.Warnf("Preload: Failed to write image %s: %v", webpFilename, err)
            continue
        }
        options := &webp.Options{Lossless: false, Quality: 82}
        if err := webp.Encode(outFile, img, options); err != nil {
            outFile.Close()
            log.Warnf("Preload: WebP encode failed for %s: %v", name, err)
            continue
        }
        outFile.Close()

        slide := Slide{
            ID:        id,
            Title:     "Preloaded: " + name,
            SourceURL: "",
            ImageURL:  "/serve?id=" + id,
            CreatedAt: time.Now().UTC(),
        }
        storeMu.Lock()
        slideStore[id] = slide
        storeMu.Unlock()
        log.Infof("Preload: Added image %s as slide ID=%s", name, id)
        count++
    }
    log.Infof("Preload: Completed - %d images loaded from %s", count, preloadDir)
}

