// Package storage defines VOID's Data Access Layer. The Simulation Core
// only ever depends on these interfaces, never on a concrete database
// driver - concrete PostgreSQL/Redis/NATS/Kafka/S3 backends implement them
// as swappable Providers (see memory.go / file.go for the bundled,
// dependency-free implementations; production deployments wire in real
// drivers behind the same interfaces).
package storage

import (
	"context"
	"time"

	"void/internal/event"
)

// KeyValueStore is the abstraction over Redis-like runtime/cache state.
type KeyValueStore interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Document is a generic record persisted to the relational/metadata store.
type Document map[string]interface{}

// MetadataStore is the abstraction over PostgreSQL-like structured storage
// for Worlds, Simulations, Scenarios, Snapshots metadata, Experiments, etc.
type MetadataStore interface {
	Put(ctx context.Context, collection, id string, doc Document) error
	Get(ctx context.Context, collection, id string) (Document, bool, error)
	List(ctx context.Context, collection string) ([]Document, error)
	Delete(ctx context.Context, collection, id string) error
}

// EventStream is the abstraction over NATS/Kafka-like event streaming.
type EventStream interface {
	Publish(ctx context.Context, topic string, e event.Event) error
	Subscribe(ctx context.Context, topic string, handler func(event.Event)) (unsubscribe func(), err error)
}

// ObjectStore is the abstraction over S3-compatible storage for large
// Snapshots and Datasets.
type ObjectStore interface {
	Put(ctx context.Context, bucket, key string, data []byte) error
	Get(ctx context.Context, bucket, key string) ([]byte, bool, error)
	List(ctx context.Context, bucket, prefix string) ([]string, error)
	Delete(ctx context.Context, bucket, key string) error
}

// Provider bundles all four backends so the rest of VOID takes a single
// dependency and callers can swap in real drivers (postgres/redis/nats/s3)
// behind these same interfaces without touching engine code.
type Provider struct {
	KV     KeyValueStore
	Meta   MetadataStore
	Stream EventStream
	Object ObjectStore
}
