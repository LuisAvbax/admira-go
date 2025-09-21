package main

import (
	"log"
	"os"
	"time"

	api "admira/internal/http"
	"admira/internal/ingest"
	"admira/internal/store"
)

func main() {
	ads := getenv("ADS_API_URL", "https://mocki.io/v1/9dcc2981-2bc8-465a-bce3-47767e1278e6")
	crm := getenv("CRM_API_URL", "https://mocki.io/v1/6a064f10-829d-432c-9f0d-24d5b8cb71c7")

	st := store.NewMemory()
	httpClient := ingest.NewHTTPClient(5 * time.Second)
	etl := &ingest.ETL{ADSURL: ads, CRMURL: crm, HTTP: httpClient, Store: st}

	srv := api.NewServer(etl, st)
	log.Println("listening on :" + getenv("PORT", "8080"))
	log.Fatal(srv.Serve())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
