package handlers

import (
    "embed"
    "fmt"
    "io/fs"
    "net/http"
    "strings"
    "time"
)

var expiry = time.Date(2026, 4, 30, 23, 59, 0, 0, time.UTC)

func RegisterStatic(mux *http.ServeMux, embedded embed.FS, apiBase string) {
    sub, err := fs.Sub(embedded, "frontend/dist")
    if err != nil {
        panic(err)
    }
    files := http.FileServer(http.FS(sub))

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if time.Now().UTC().After(expiry) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.WriteHeader(http.StatusForbidden)
            expiryText := expiry.UTC().Format("January 2, 2006 at 15:04 GMT")
            html := fmt.Sprintf(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Expired</title>
</head>
<body style="font-family:sans-serif;padding:40px">
  <h1>Application expired</h1>
  <p>This build expired after %s.</p>
</body>
</html>`, expiryText)
            _, _ = w.Write([]byte(html))
            return
        }

        path := strings.TrimPrefix(r.URL.Path, "/")
        if path == "" || path == "index.html" {
            b, err := fs.ReadFile(sub, "index.html")
            if err != nil {
                http.Error(w, "index not found", http.StatusInternalServerError)
                return
            }
            html := strings.Replace(string(b), "__API_BASE__", apiBase, 1)
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            _, _ = w.Write([]byte(html))
            return
        }

        if _, err := fs.Stat(sub, path); err == nil {
            files.ServeHTTP(w, r)
            return
        }

        // SPA fallback.
        b, err := fs.ReadFile(sub, "index.html")
        if err != nil {
            http.Error(w, fmt.Sprintf("index not found: %v", err), http.StatusInternalServerError)
            return
        }
        html := strings.Replace(string(b), "__API_BASE__", apiBase, 1)
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        _, _ = w.Write([]byte(html))
    })
}
