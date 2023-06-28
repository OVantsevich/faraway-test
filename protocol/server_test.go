package protocol

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestServer_Serve(t *testing.T) {
	logger, _ := zapLoggerInit("test")
	server := NewServer(logger.Sugar(), NewProofOfWork(20, time.Second*60), time.Second*60, func(request Request) (Response, error) {
		return "Test quote", nil
	})
	l, _ := net.Listen("tcp", "localhost:12345")
	go server.Serve(l)

	for i := 0; i < 10; i++ {
		go func() {
			conn, _ := net.Dial("tcp", "localhost:12345")
			c, err := NewClient(conn)
			require.NoError(t, err)
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
			fmt.Print(c.GetQuote())
		}()
	}
	time.Sleep(time.Second * 20)
}

func zapLoggerInit(serviceName string) (*zap.Logger, error) {
	srvField := zap.Fields(zap.Field{
		Key:    "service",
		Type:   zapcore.StringType,
		String: serviceName,
	})

	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build(srvField)
}
