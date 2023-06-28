package protocol

import (
	"math/rand"
	"net"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/require"
)

func TestServer_ServePOW(t *testing.T) {
	logger, _ := zapLoggerInit("test")
	address := "localhost:12345"
	testQuote := "Test quote"

	server := NewServer(logger.Sugar(), NewProofOfWork(20, time.Second*60), time.Second*60, func(request *Request) (*Response, error) {
		response := Response(testQuote)
		return &response, nil
	})
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)
	go server.Serve(l)

	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	c, err := NewClient(conn)
	require.NoError(t, err)
	quote, err := c.GetQuote()
	require.NoError(t, err)
	require.Equal(t, testQuote, quote)

	for i := 0; i < 10; i++ {
		data := make([]byte, i)
		for j := range data {
			data[j] = letter[rand.Intn(len(letter))]
		}
		testQuote = string(data)
		quote, err = c.GetQuote()
		require.NoError(t, err)
		require.Equal(t, testQuote, quote)
	}
}

func TestServer_Serve(t *testing.T) {
	logger, _ := zapLoggerInit("test")
	address := "localhost:12344"
	testQuote := "Test quote"

	server := NewServer(logger.Sugar(), nil, time.Second*60, func(request *Request) (*Response, error) {
		response := Response(testQuote)
		return &response, nil
	})
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)
	go server.Serve(l)

	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	c, err := NewClient(conn)
	require.NoError(t, err)
	quote, err := c.GetQuote()
	require.NoError(t, err)
	require.Equal(t, testQuote, quote)

	for i := 0; i < 10; i++ {
		data := make([]byte, i)
		for j := range data {
			data[j] = letter[rand.Intn(len(letter))]
		}
		testQuote = string(data)
		quote, err = c.GetQuote()
		require.NoError(t, err)
		require.Equal(t, testQuote, quote)
	}
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
