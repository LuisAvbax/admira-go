package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type ctxKey int
const requestIDKey ctxKey = 1

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		rid := time.Now().UnixNano() ^ rand.Int63()
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		w.Header().Set("X-Request-ID",  fmt.Sprintf("%x", rid))
		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		logJSON(map[string]any{
			"level":"info","msg":"request",
			"method":r.Method,"path":r.URL.Path,"status":"ok",
			"duration_ms": time.Since(start).Milliseconds(),
			"request_id": fmt.Sprintf("%x", rid),
		})
	})
}

func logJSON(v any) {
	b, _ := json.Marshal(v)
	log.Println(string(b))
}
