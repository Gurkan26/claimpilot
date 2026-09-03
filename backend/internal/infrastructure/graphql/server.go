package graphql

import (
	"net/http"

	"log/slog"
)

// NewServer creates an HTTP handler for the GraphQL endpoint.
// This is a placeholder that will be fully wired after gqlgen code generation.
// For now, it serves a simple introspection response.
func NewServer(resolver *Resolver, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// GraphQL endpoint (placeholder — will be replaced by gqlgen handler)
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"__typename":"Query"},"extensions":{"message":"ClaimPilot GraphQL API - gqlgen integration pending"}}`))
	})

	// GraphQL Playground (development only)
	mux.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>ClaimPilot GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root"></div>
  <script>
    window.addEventListener('load', function() {
      GraphQLPlayground.init(document.getElementById('root'), { endpoint: '/graphql' })
    })
  </script>
</body>
</html>`))
	})

	logger.Info("GraphQL server initialized",
		"endpoints", []string{"/graphql", "/playground"},
	)

	return mux
}
