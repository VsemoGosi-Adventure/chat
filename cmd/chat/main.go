package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/VsemoGosi-Adventure/chat/internal/config"
	"github.com/VsemoGosi-Adventure/chat/internal/db"
	"github.com/VsemoGosi-Adventure/chat/internal/handler"
	"github.com/VsemoGosi-Adventure/chat/internal/middleware"
	"github.com/VsemoGosi-Adventure/chat/internal/service"
	"github.com/VsemoGosi-Adventure/chat/internal/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application failed: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()

	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse database DSN: %w", err)
	}

	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = cfg.Database.MinConns
	poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.Database.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("database connection established")

	queries := db.New(pool)
	chatService := service.NewChatService(queries)

	hub := websocket.NewHub(cfg.WebSocket.BroadcastBuffer)
	hubCtx, hubCancel := context.WithCancel(ctx)
	defer hubCancel()

	go hub.Run(hubCtx)

	chatHandler := handler.NewChatHandler(chatService)
	wsHandler := handler.NewWebSocketHandler(hub, chatService, cfg.WebSocket)

	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/chats/{event_id}/messages", middleware.AuthMiddleware(http.HandlerFunc(chatHandler.GetMessages)))
	mux.Handle("GET /api/v1/chats/{event_id}/ws", middleware.AuthMiddleware(http.HandlerFunc(wsHandler.HandleWebSocket)))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Printf("received signal %v, starting graceful shutdown", sig)

		hubCancel()

		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			srv.Close()
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		log.Println("server stopped gracefully")
	}

	return nil
}
