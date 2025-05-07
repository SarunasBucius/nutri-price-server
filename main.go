package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/SarunasBucius/nutri-price-server/graph"
	"github.com/SarunasBucius/nutri-price-server/internal/setup"
	"github.com/SarunasBucius/nutri-price-server/migrations"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	ctx := context.Background()

	config, err := setup.LoadConfig(ctx)
	if err != nil {
		slog.Error("load config", "error", err)
		return
	}

	if err := migrations.MigrateDB(ctx, config.DBPool); err != nil {
		slog.Error("migrate db", "error", err)
		return
	}

	r := setup.LoadRouter(config)

	slog.InfoContext(ctx, "Listening...", "port", config.Port)

	port := config.Port
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	attachGraphQLRoutes(config.DBPool, r, config.DynamoDB)

	if err := http.ListenAndServe(port, r); err != nil {
		slog.Error("listen and serve", "error", err)
		return
	}
}

func attachGraphQLRoutes(db *pgxpool.Pool, r *chi.Mux, dynamoDB *dynamodb.Client) {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		DB:       db,
		DynamoDB: dynamoDB,
	}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	r.Handle("/", playground.Handler("GraphQL playground", "/query"))
	r.Handle("/query", srv)
}
