package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"admira/internal/ingest"
	"admira/internal/metrics"
	"admira/internal/store"
)

type Server struct {
	mux  *http.ServeMux
	etl  *ingest.ETL
	st   *store.Memory
}

func NewServer(etl *ingest.ETL, st *store.Memory) *Server {
	s := &Server{mux: http.NewServeMux(), etl: etl, st: st}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(200); w.Write([]byte("ok")) })
	s.mux.HandleFunc("/readyz",  func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(200); w.Write([]byte("ready")) })

	s.mux.HandleFunc("/ingest/run", s.handleIngestRun)
	s.mux.HandleFunc("/metrics/channel", s.handleMetricsChannel)
	// s.mux.HandleFunc("/metrics/funnel", s.handleMetricsFunnel) // análogo

	// (Opcional) s.mux.HandleFunc("/export/run", s.handleExportRun)
}

func (s *Server) Serve() error {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	return http.ListenAndServe(":"+port, withRequestID(s.mux))
}

func (s *Server) handleIngestRun(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("since")
	var sincePtr *time.Time
	if q != "" {
		if t, err := time.Parse("2006-01-02", q); err == nil {
			sincePtr = &t
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second); defer cancel()
	err := s.etl.Run(ctx, sincePtr)
	if err != nil { http.Error(w, err.Error(), 502); return }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status":"ok"})
}

func (s *Server) handleMetricsChannel(w http.ResponseWriter, r *http.Request) {
	// filtros
	var f metrics.Filters
	if v := r.URL.Query().Get("from"); v != "" { t,_ := time.Parse("2006-01-02", v); f.From = &t }
	if v := r.URL.Query().Get("to");   v != "" { t,_ := time.Parse("2006-01-02", v); f.To   = &t }
	f.Channel    = r.URL.Query().Get("channel")
	f.Campaign   = r.URL.Query().Get("campaign_id")
	f.UTMCampaign= r.URL.Query().Get("utm_campaign")
	f.UTMSource  = r.URL.Query().Get("utm_source")
	f.UTMMedium  = r.URL.Query().Get("utm_medium")
	f.Limit, _   = strconv.Atoi(r.URL.Query().Get("limit"))
	f.Offset, _  = strconv.Atoi(r.URL.Query().Get("offset"))

	rows := metrics.ChannelMetrics(s.st.AllAds(), s.st.AllCRM(), f)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count": len(rows),
		"items": rows,
	})
}
