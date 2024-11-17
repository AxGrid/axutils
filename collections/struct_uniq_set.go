package collections

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/gob"
	"sync"
	"time"
)

type structUniqSetElement struct {
	hash      [16]byte
	createdAt time.Time
}

type StructUniqSet[V any] struct {
	build         func(v V, keys ...any) ([16]byte, error)
	elements      map[[16]byte]*structUniqSetElement
	ttlElements   []*structUniqSetElement
	elementsMu    sync.RWMutex
	ctx           context.Context
	ttl           time.Duration
	checkInterval time.Duration
}

type StructUniqSetBuilder[V any] struct {
	build         func(v V, keys ...any) ([16]byte, error)
	ctx           context.Context
	ttl           time.Duration
	checkInterval time.Duration
}

func NewStructUniqSet[V any]() *StructUniqSetBuilder[V] {
	return &StructUniqSetBuilder[V]{
		ctx:           context.Background(),
		ttl:           time.Minute * 30,
		checkInterval: time.Second * 1,
		build: func(v V, keys ...any) ([16]byte, error) {
			return getHashCode(v)
		},
	}
}

func (b *StructUniqSetBuilder[V]) WithBuildFunc(buildFn func(v V, keys ...any) ([16]byte, error)) *StructUniqSetBuilder[V] {
	b.build = buildFn
	return b
}

func (b *StructUniqSetBuilder[V]) WithContext(ctx context.Context) *StructUniqSetBuilder[V] {
	b.ctx = ctx
	return b
}

func (b *StructUniqSetBuilder[V]) WithElementTtl(ttl time.Duration) *StructUniqSetBuilder[V] {
	b.ttl = ttl
	return b
}

func (b *StructUniqSetBuilder[V]) WithCheckInterval(interval time.Duration) *StructUniqSetBuilder[V] {
	b.checkInterval = interval
	return b
}

func (b *StructUniqSetBuilder[V]) Build() *StructUniqSet[V] {
	res := &StructUniqSet[V]{
		build:         b.build,
		elements:      map[[16]byte]*structUniqSetElement{},
		ctx:           b.ctx,
		ttl:           b.ttl,
		checkInterval: b.checkInterval,
	}
	go func() {
		for {
			select {
			case <-res.ctx.Done():
				return
			case <-time.After(res.checkInterval):
				res.elementsMu.RLock()
				first := res.ttlElements[0]
				res.elementsMu.RUnlock()
				if time.Since(first.createdAt) < res.ttl {
					continue
				}
				res.elementsMu.Lock()
				for i, el := range res.ttlElements {
					if time.Since(el.createdAt) < res.ttl {
						res.ttlElements = res.ttlElements[i:]
						break
					}
					delete(res.elements, el.hash)
				}
				res.elementsMu.Unlock()
			}
		}
	}()
	return res
}

func (s *StructUniqSet[V]) getHashCode(v V) ([16]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(v)
	if err != nil {
		return [16]byte{}, err
	}
	md5.Sum(buffer.Bytes())
	return md5.Sum(buffer.Bytes()), nil
}

func (s *StructUniqSet[V]) Count() int {
	s.elementsMu.RLock()
	defer s.elementsMu.RUnlock()
	return len(s.elements)
}

func (s *StructUniqSet[V]) Has(v V) (bool, error) {
	hash, err := s.getHashCode(v)
	if err != nil {
		return false, err
	}
	s.elementsMu.RLock()
	_, ok := s.elements[hash]
	s.elementsMu.RUnlock()
	return ok, nil
}

func (s *StructUniqSet[V]) Add(v V) (bool, error) {
	hash, err := s.getHashCode(v)
	if err != nil {
		return false, err
	}
	s.elementsMu.RLock()
	_, ok := s.elements[hash]
	s.elementsMu.RUnlock()
	if ok {
		return false, nil
	}
	s.elementsMu.Lock()
	_, ok = s.elements[hash]
	if !ok {
		s.elements[hash] = &structUniqSetElement{hash: hash, createdAt: time.Now()}
		s.ttlElements = append(s.ttlElements, s.elements[hash])
	}
	s.elementsMu.Unlock()
	return !ok, nil
}

func (s *StructUniqSet[V]) Remove(v V) (bool, error) {
	hash, err := s.getHashCode(v)
	if err != nil {
		return false, err
	}
	s.elementsMu.Lock()
	_, ok := s.elements[hash]
	if ok {
		delete(s.elements, hash)
	}
	s.elementsMu.Unlock()
	return ok, nil
}

func getHashCode(v any) ([16]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(v)
	if err != nil {
		return [16]byte{}, err
	}
	md5.Sum(buffer.Bytes())
	return md5.Sum(buffer.Bytes()), nil
}
