package httpx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Veysel440/go-audit/internal/core"
	"github.com/Veysel440/go-audit/internal/repo/mongo"
	"github.com/Veysel440/go-audit/internal/service"
	"github.com/Veysel440/go-audit/pkg/rate"
)

func NewMux(logger *slog.Logger) (http.Handler, func(context.Context) error) {
	// config
	mongoURI := Getenv("MONGO_URI", "mongodb://localhost:27017")
	db := Getenv("MONGO_DB", "Go")
	col := Getenv("MONGO_COLLECTION", "Go")
	retentionDays := int64(GetInt("RETENTION_DAYS", 90))
	apiKeys := Getenv("API_KEYS", "") // "k1,k2,..."
	rps := GetInt("RATE", 60)

	repo, err := mongo.New(context.Background(), mongo.Config{
		URI: mongoURI, DB: db, Collection: col, RetentionDays: retentionDays,
	})
	if err != nil {
		panic(err)
	}

	svc := service.New(repo)
	mux := http.NewServeMux()

	// health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
	})

	// ingest
	mux.HandleFunc("POST /v1/audits", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			ActorID, ActorType, Action, ResourceID, ResourceType string
			IP, UA                                               string
			Metadata                                             map[string]any
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid_json")
			return
		}

		idem := r.Header.Get("Idempotency-Key")
		hash := ""
		if idem != "" {
			h := sha256.Sum256([]byte(idem))
			hash = hex.EncodeToString(h[:])
		}

		out, err := svc.Create(r.Context(), core.CreateAudit{
			ActorID: in.ActorID, ActorType: in.ActorType, Action: in.Action,
			ResourceID: in.ResourceID, ResourceType: in.ResourceType,
			IP: in.IP, UA: in.UA, Metadata: in.Metadata, IdemHash: hash,
		})
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, out)
	})

	// get
	mux.HandleFunc("GET /v1/audits/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		out, err := svc.Get(r.Context(), id)
		if err != nil {
			writeErr(w, http.StatusNotFound, "not_found")
			return
		}
		writeJSON(w, http.StatusOK, out)
	})

	// list
	mux.HandleFunc("GET /v1/audits", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var f service.ListFilter
		f.ActorID = q.Get("actorId")
		f.ActorType = q.Get("actorType")
		f.Action = q.Get("action")
		f.ResourceID = q.Get("resourceId")
		f.ResourceType = q.Get("resourceType")
		if v := q.Get("since"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				f.Since = &t
			}
		}
		if v := q.Get("until"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				f.Until = &t
			}
		}
		if v := q.Get("limit"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				f.Limit = n
			}
		}
		if v := q.Get("offset"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				f.Offset = n
			}
		}

		list, err := svc.List(r.Context(), f)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad_request")
			return
		}
		writeJSON(w, http.StatusOK, list)
	})

	auth := NewAPIKeyAuth(apiKeys)
	rl := rate.New(rps, time.Minute)

	handler := withRecover(withLogging(logger, auth.Middleware(withRate(rl)(mux))))

	return handler, func(ctx context.Context) error { return nil }
}

// Exported helpers
func Getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func GetInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}
