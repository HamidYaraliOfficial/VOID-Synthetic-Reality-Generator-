package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"void/internal/event"
)

// ---- In-memory KeyValueStore (stands in for Redis in dev/single-node) ----

type kvEntry struct {
	value   string
	expires time.Time
}

type MemoryKV struct {
	mu   sync.RWMutex
	data map[string]kvEntry
}

func NewMemoryKV() *MemoryKV { return &MemoryKV{data: map[string]kvEntry{}} }

func (m *MemoryKV) Get(_ context.Context, key string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[key]
	if !ok {
		return "", false, nil
	}
	if !e.expires.IsZero() && time.Now().After(e.expires) {
		return "", false, nil
	}
	return e.value, true, nil
}

func (m *MemoryKV) Set(_ context.Context, key, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.data[key] = kvEntry{value: value, expires: exp}
	return nil
}

func (m *MemoryKV) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// ---- In-memory MetadataStore (stands in for PostgreSQL in dev) ----

type MemoryMeta struct {
	mu   sync.RWMutex
	data map[string]map[string]Document // collection -> id -> doc
}

func NewMemoryMeta() *MemoryMeta { return &MemoryMeta{data: map[string]map[string]Document{}} }

func (m *MemoryMeta) Put(_ context.Context, collection, id string, doc Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data[collection] == nil {
		m.data[collection] = map[string]Document{}
	}
	m.data[collection][id] = doc
	return nil
}

func (m *MemoryMeta) Get(_ context.Context, collection, id string) (Document, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.data[collection]
	if !ok {
		return nil, false, nil
	}
	d, ok := c[id]
	return d, ok, nil
}

func (m *MemoryMeta) List(_ context.Context, collection string) ([]Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c := m.data[collection]
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Document, 0, len(ids))
	for _, id := range ids {
		out = append(out, c[id])
	}
	return out, nil
}

func (m *MemoryMeta) Delete(_ context.Context, collection, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.data[collection]; ok {
		delete(c, id)
	}
	return nil
}

// ---- In-process EventStream (stands in for NATS/Kafka in dev) ----

type MemoryStream struct {
	mu   sync.RWMutex
	subs map[string][]func(event.Event)
}

func NewMemoryStream() *MemoryStream { return &MemoryStream{subs: map[string][]func(event.Event){}} }

func (s *MemoryStream) Publish(_ context.Context, topic string, e event.Event) error {
	s.mu.RLock()
	handlers := append([]func(event.Event){}, s.subs[topic]...)
	s.mu.RUnlock()
	for _, h := range handlers {
		go h(e)
	}
	return nil
}

func (s *MemoryStream) Subscribe(_ context.Context, topic string, handler func(event.Event)) (func(), error) {
	s.mu.Lock()
	s.subs[topic] = append(s.subs[topic], handler)
	idx := len(s.subs[topic]) - 1
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if idx < len(s.subs[topic]) {
			s.subs[topic][idx] = nil
		}
	}, nil
}

// ---- Filesystem-backed ObjectStore (stands in for S3 in dev) ----

type FileObjectStore struct {
	mu   sync.Mutex
	root string
}

func NewFileObjectStore(root string) *FileObjectStore {
	return &FileObjectStore{root: root}
}

func (f *FileObjectStore) path(bucket, key string) string {
	return fmt.Sprintf("%s/%s/%s", f.root, bucket, key)
}

func (f *FileObjectStore) Put(_ context.Context, bucket, key string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return writeFileAll(f.path(bucket, key), data)
}

func (f *FileObjectStore) Get(_ context.Context, bucket, key string) ([]byte, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, ok, err := readFile(f.path(bucket, key))
	return data, ok, err
}

func (f *FileObjectStore) List(_ context.Context, bucket, prefix string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listFiles(fmt.Sprintf("%s/%s", f.root, bucket), prefix)
}

func (f *FileObjectStore) Delete(_ context.Context, bucket, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return deleteFile(f.path(bucket, key))
}

// NewDefaultProvider wires the bundled in-memory/filesystem implementations
// - suitable for local dev and small simulations out of the box. Point
// PROVIDER=postgres|redis|nats|s3 env vars (see config package) at a real
// deployment to swap in production drivers behind the same interfaces.
func NewDefaultProvider(dataDir string) *Provider {
	return &Provider{
		KV:     NewMemoryKV(),
		Meta:   NewMemoryMeta(),
		Stream: NewMemoryStream(),
		Object: NewFileObjectStore(strings.TrimSuffix(dataDir, "/") + "/objects"),
	}
}
