package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"time"

	"github.com/OVantsevich/faraway-test/protocol"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/OVantsevich/faraway-test/server/infrastructure/logger"
	"github.com/OVantsevich/faraway-test/server/internal/config"
	"github.com/OVantsevich/faraway-test/server/internal/ent"
	"github.com/OVantsevich/faraway-test/server/internal/handler"
	"github.com/OVantsevich/faraway-test/server/internal/migrations"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		stdlog.Fatal(err)
	}

	zapLogger, err := zapLoggerInit(cfg.Environment, cfg.ServiceName)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	client, err := ent.Open("sqlite3", cfg.SqliteConn())
	if err != nil {
		logger.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		logger.Fatalf("failed creating schema resources: %v", err)
	}
	client, err = migrations.QuoteMigrations(ctx, client)
	if err != nil {
		logger.Fatalf("failed migrating schema resources: %v", err)
	}

	quoteHandler := handler.NewQuoteHandler(client, logger)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort))
	server := protocol.NewServer(logger, nil, time.Second*120, quoteHandler.GetQuote)

	logger.Infof("Server listened on: %v", l.Addr())
	logger.Fatal(server.Serve(l))
}

func zapLoggerInit(env config.Environment, serviceName string) (*zap.Logger, error) {
	srvField := zap.Fields(zap.Field{
		Key:    "service",
		Type:   zapcore.StringType,
		String: serviceName,
	})

	if env == config.Production {
		return logger.Production(srvField)
	}

	if env == config.Develop {
		return logger.Development(srvField)
	}

	return logger.Production(srvField)
}
