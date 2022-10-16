package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var appName = "myapp"

var projectID = os.Getenv("PROJECT_ID")
var servicePort = os.Getenv("PORT")

func main() {

	oplog := httplog.LogEntry(context.Background())
	/* jsonify logging */
	httpLogger := httplog.NewLogger(appName, httplog.Options{JSON: true, LevelFieldName: "severity", Concise: true})

	/* exporter for prometheus */
	m := chiprometheus.NewMiddleware(appName)

	v := chi.NewRouter()
	// r.Use(middleware.Logger)
	v.Use(middleware.RequestID)
	v.Use(middleware.Timeout(60 * time.Second))
	v.Use(httplog.RequestLogger(httpLogger))
	v.Use(m)

	v.Handle("/metrics", promhttp.Handler())

	v.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]string{"Ping": "Pong"})
	})

	v.Get("/api/author/{user:[a-z0-9-.]+}", getFirestore)

	v2 := chi.NewRouter()
	v2.Use(m)
	v2.Use(middleware.Timeout(60 * time.Second))
	v2.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":10080", v2); err != nil {
			oplog.Err(err)
		}
	}()

	if err := http.ListenAndServe(":"+servicePort, v); err != nil {
		oplog.Err(err)
	}

}

func getFirestore(w http.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())
	ctx := r.Context()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		oplog.Err(err)
		return
	}
	username := chi.URLParam(r, "user")
	/* trick to get just one record */
	query := client.Collection("authors").Where("username", "==", username).Limit(1)
	itr := query.Documents(ctx)
	defer itr.Stop()

	snap, err := itr.Next()
	if err != nil {
		oplog.Err(err)
	}

	render.JSON(w, r, snap.Data())

}
