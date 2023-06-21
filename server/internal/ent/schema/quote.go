// Package schema define ent schemes
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Quote holds the schema definition for the Quote entity
type Quote struct {
	ent.Schema
}

// Fields of the Quote
func (Quote) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Default(uuid.New().String()).
			StorageKey("oid"),
		field.String("data"),
		field.Time("created"),
		field.Time("updated"),
	}
}

// Indexes of the Quote
func (Quote) Indexes() []ent.Index {
	return []ent.Index{}
}
