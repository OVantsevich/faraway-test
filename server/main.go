package main

import (
	"bufio"
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	stdlog "log"
	"net"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/OVantsevich/faraway-test/server/infrastructure/logger"
	"github.com/OVantsevich/faraway-test/server/internal/config"
	"github.com/OVantsevich/faraway-test/server/internal/ent"
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

	_, err = client.Quote.Create().SetID("dsa").SetData("ds").SetCreated(time.Now()).SetUpdated(time.Now()).Save(context.Background())
	if err != nil {
		logger.Error(err)
	}
	q, err := client.Quote.Query().All(context.Background())
	if err != nil {
		logger.Error(err)
	}
	fmt.Print(q)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort))
	c, err := l.Accept()

	r := bufio.NewReader(c)
	r.ReadString('a')

	http.Post()

	time.Sleep(time.Hour)
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
