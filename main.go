package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"investment-analysis/handlers"
	"investment-analysis/persistence"
	"investment-analysis/persistence/postgres"
	"investment-analysis/persistence/sqlite"
)

//go:embed frontend/dist/*
var embeddedUI embed.FS

const version = "0.0.5"

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	_ = exec.Command(cmd, args...).Start()
}

func main() {
	uiPort := flag.Int("ui-port", 8080, "Port for the embedded web UI")
	apiPort := flag.Int("api-port", 3000, "Port for the JSON API")
	storeBackend := flag.String("store-backend", "sqlite", "Storage backend: sqlite or postgres")
	noBrowser := flag.Bool("no-browser", false, "Disable automatic browser launch")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	var (
		db  persistence.Store
		err error
	)

	switch *storeBackend {
	case "postgres":
		db, err = postgres.Open()
	case "sqlite":
		db, err = sqlite.Open()
	default:
		log.Fatalf("unsupported store backend %q", *storeBackend)
	}
	if err != nil {
		log.Fatalf("open %s store: %v", *storeBackend, err)
	}
	defer db.Close()

	apiAddr := fmt.Sprintf(":%d", *apiPort)
	uiAddr := fmt.Sprintf(":%d", *uiPort)
	_ = fmt.Sprintf("http://localhost:%d", *apiPort)
	uiURL := fmt.Sprintf("http://localhost:%d", *uiPort)

	go func() {
		apiMux := http.NewServeMux()
		handlers.RegisterAPI(apiMux, db)
		log.Printf("API listening on %s", apiAddr)
		if err := http.ListenAndServe(apiAddr, apiMux); err != nil {
			log.Fatalf("API server stopped: %v", err)
		}
	}()

	uiMux := http.NewServeMux()
	handlers.RegisterAPI(uiMux, db)
	handlers.RegisterStatic(uiMux, embeddedUI, "/api")
	log.Printf("UI listening on %s", uiAddr)

	if !*noBrowser {
		openBrowser(uiURL)
	}

	if err := http.ListenAndServe(uiAddr, uiMux); err != nil {
		log.Fatalf("UI server stopped: %v", err)
	}
}
