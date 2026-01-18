package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"image-api/config"

	"github.com/google/uuid"
)

type ImageResponse struct {
    URL      string `json:"url"`
    Filename string `json:"filename"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

func main() {
    if err := config.Load(); err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    log.Printf("Loaded %d API keys", len(config.Configuration.APIKeys))

    if err := os.MkdirAll("./images", 0755); err != nil {
        log.Fatal("Failed to create upload directory:", err)
    }

    http.HandleFunc("/new", authMiddleware(uploadHandler))
    http.HandleFunc("/images/", serveImageHandler)
    http.HandleFunc("/ping", pingHandler)

    log.Printf("Server starting on port %s", config.Configuration.Port)
    log.Fatal(http.ListenAndServe(
        ":"+config.Configuration.Port, 
        http.DefaultServeMux,
    ))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        
        if apiKey == "" {
            respondWithError(w, http.StatusUnauthorized, 
                "Missing API key")
            return
        }

        if !config.Configuration.IsValidAPIKey(apiKey) {
            respondWithError(w, http.StatusUnauthorized, 
                "Invalid API key")
            return
        }

        next(w, r)
    }
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        respondWithError(w, http.StatusMethodNotAllowed, 
            "Method not allowed")
        return
    }

    if r.ContentLength > config.Configuration.MaxFileSize {
        respondWithError(w, http.StatusBadRequest, 
            fmt.Sprintf("File too large. Max size: %dMB", 
                config.Configuration.MaxFileSize/(1<<20)))
        return
    }

    contentType := r.Header.Get("Content-Type")
    if !isValidImageType(contentType) {
        respondWithError(w, http.StatusBadRequest, 
            "Invalid content type. Only png, jpg, gif and webp allowed")
        return
    }

    ext := getExtensionFromContentType(contentType)
    if ext == "" {
        respondWithError(w, http.StatusBadRequest, 
            "Unsupported image type. Only png, jpg, gif and webp allowed")
        return
    }

    uniqueID := uuid.New().String()
    filename := fmt.Sprintf("%s%s", uniqueID, ext)
    filepath := filepath.Join("./images", filename)

    dst, err := os.Create(filepath)
    if err != nil {
		log.Println("error > " + err.Error())
        respondWithError(w, http.StatusInternalServerError, 
            "Failed to save file")
        return
    }
    defer dst.Close()

    limitedReader := io.LimitReader(r.Body, config.Configuration.MaxFileSize)
    written, err := io.Copy(dst, limitedReader)
    if err != nil {
        os.Remove(filepath)
        respondWithError(w, http.StatusInternalServerError, 
            "Failed to save file")
        return
    }

    if written == config.Configuration.MaxFileSize {
        os.Remove(filepath)
        respondWithError(w, http.StatusBadRequest, 
            fmt.Sprintf("File too large. Max size: %dMB", 
                config.Configuration.MaxFileSize/(1<<20)))
        return
    }

    imageURL := fmt.Sprintf("/images/%s", filename)
    response := ImageResponse{
        URL:      imageURL,
        Filename: filename,
    }
	log.Println("Recieved and saved image " + filename)

	res, _ := json.Marshal(response)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    w.Write(res)
}

func serveImageHandler(w http.ResponseWriter, r *http.Request) {
    filename := strings.TrimPrefix(r.URL.Path, "/images/")
    if filename == "" {
        respondWithError(w, http.StatusBadRequest, "Invalid image path")
        return
    }

	split := strings.Split(filename, ".")
	if (len(split) != 2) {
        respondWithError(w, http.StatusBadRequest, "Invalid image path")
		return
	}

	err := uuid.Validate(split[0])
	if (err != nil) {
        respondWithError(w, http.StatusBadRequest, "Invalid image path")
		return;
	}

    filepath := filepath.Join("./images", filename)
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        respondWithError(w, http.StatusNotFound, "Image not found")
        return
    }

	log.Println("Returned image " + filename)

    http.ServeFile(w, r, filepath)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func isValidImageType(contentType string) bool {
    validTypes := []string{
        "image/jpeg",
        "image/png",
        "image/gif",
        "image/webp",
    }
	return slices.Contains(validTypes, contentType)
}

func getExtensionFromContentType(contentType string) string {
    extensions := map[string]string{
        "image/jpeg": ".jpg",
        "image/png":  ".png",
        "image/gif":  ".gif",
        "image/webp": ".webp",
    }
    return extensions[contentType]
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	res, _ := json.Marshal(ErrorResponse{Error: message})
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(res)
}
