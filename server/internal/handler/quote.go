package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"go.uber.org/zap"
	"math/big"
	"strconv"

	"github.com/OVantsevich/faraway-test/protocol"

	"github.com/OVantsevich/faraway-test/server/internal/ent"
)

// Quote handler
type Quote struct {
	client *ent.Client
	logger *zap.SugaredLogger
}

func NewQuoteHandler(client *ent.Client, logger *zap.SugaredLogger) *Quote {
	return &Quote{client: client, logger: logger}
}

// GetQuote - receiving random quote
func (s *Quote) GetQuote(_ *protocol.Request) (*protocol.Response, error) {
	ids, err := s.client.Quote.Query().IDs(context.Background())
	if err != nil {
		return nil, fmt.Errorf("GetQuote - IDs: %v", err)
	}

	rnd, err := rand.Int(rand.Reader, big.NewInt(int64(len(ids))))
	if err != nil {
		return nil, fmt.Errorf("GetQuote - Int: %v", err)
	}

	quote, err := s.client.Quote.Get(context.Background(), ids[rnd.Int64()])
	if err != nil {
		return nil, fmt.Errorf("GetQuote - Get: %v. ID: %v", err, strconv.Itoa(int(rnd.Int64())))
	}

	response := protocol.Response(quote.Data)
	return &response, nil
}
