package migrations

import (
	"context"
	"github.com/OVantsevich/faraway-test/server/internal/ent/quote"
	"time"

	"github.com/OVantsevich/faraway-test/server/internal/ent"
)

var quoteData = [5][2]string{
	{"0", "You create your own opportunities. Success doesn’t just come and find you–you have to go out and get it."},
	{"1", "Never break your promises. Keep every promise; it makes you credible."},
	{"2", "You are never as stuck as you think you are. Success is not final, and failure isn’t fatal."},
	{"3", "Happiness is a choice. For every minute you are angry, you lose 60 seconds of your own happiness."},
	{"4", "Habits develop into character. Character is the result of our mental attitude and the way we spend our time."},
}

// QuoteMigrations - ent migration hook for adding quotes quote table
func QuoteMigrations(ctx context.Context, client *ent.Client) (*ent.Client, error) {
	bulk := make([]*ent.QuoteCreate, len(quoteData))
	var total = 0
	for _, p := range quoteData {
		if ok, err := client.Quote.Query().Where(quote.ID(p[0])).Exist(ctx); !ok && err == nil {
			bulk[total] = client.Quote.
				Create().
				SetID(p[0]).
				SetData(p[1]).
				SetCreated(time.Now()).
				SetUpdated(time.Now())
			total++
		}
	}

	_, err := client.Quote.CreateBulk(bulk[0:total]...).Save(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}
