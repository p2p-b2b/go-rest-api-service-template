package service

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
)

type CacheRepository interface {
	Do(ctx context.Context, cmd valkey.Completed) (resp valkey.ValkeyResult)
	DoCache(ctx context.Context, cmd valkey.Cacheable, ttl time.Duration) (resp valkey.ValkeyResult)
	B() valkey.Builder
}

type CacheServiceConf struct {
	Cache        CacheRepository
	QueryTimeout time.Duration
	EntitiesTTL  time.Duration
}

type CacheService struct {
	cache        CacheRepository
	queryTimeout time.Duration
	entitiesTTL  time.Duration
}

func NewCacheService(conf CacheServiceConf) *CacheService {
	return &CacheService{
		cache:        conf.Cache,
		queryTimeout: conf.QueryTimeout,
		entitiesTTL:  conf.EntitiesTTL,
	}
}

// CacheEncoderType is the type of encoder to use when encoding and decoding values in cache
type CacheEncoderType string

const (
	CacheEncoderTypeJSON CacheEncoderType = "json"
	CacheEncoderTypeGob  CacheEncoderType = "gob"
)

func (CacheEncoderType CacheEncoderType) String() string {
	return string(CacheEncoderType)
}

// FromCacheOrDB gets the value from cache if it exists, otherwise it queries the value and sets it in cache
// if the query is successful. If the query fails, the error is returned.
// When cache operation fails, no error is returned, the query is executed and the value is not set in cache.
// This serializer uses JSON to encode and decode the value.
//
// ctx: context used to wait get and set operations on cache server, this must be lowed to avoid blocking
// when the cache server is slow or down
// cache: cache repository to use
// key: key to use in cache
// cacheType: type of encoder to use when encoding and decoding values in cache
// query: function to query the value if it does not exist in cache
// ttl: time to live for the value in cache
//
// Example:
// ctxCache, cancel := context.WithTimeout(ctx, ref.cacheService.queryTimeout)
//
//	defer cancel()
//
// userAuth, err = FromCacheOrDB[repository.SelectAuthzOutput](
//
//	ctxCache,
//	ref.cacheService.cache,
//	cacheKey,
//	CacheEncoderTypeJSON,
//	func() (*repository.SelectAuthzOutput, error) {
//		return ref.repository.SelectAuthz(ctx, userID)
//	}, ref.cacheService.entitiesTTL,
//
// )
//
//	if err != nil {
//		....
//	}
func FromCacheOrDB[T any](ctx context.Context, cache CacheRepository, key string, cacheType CacheEncoderType, query func() (T, error), timeout, ttl time.Duration) (T, error) {
	var value (T)

	// try to get from cache
	cacheResult := cache.DoCache(ctx, valkey.Cacheable(cache.B().Get().Key(key).Build()), timeout)

	// some value was found in cache
	if cacheResult.Error() == nil {
		// try to decode the value
		switch cacheType {
		case CacheEncoderTypeJSON:
			if cacheResult.DecodeJSON(&value) != nil {
				slog.Warn("could not decode value from cache", "cache", "error", "error", cacheResult.Error())
				return value, cacheResult.Error()
			}

			slog.Debug("service.Cache.FromCacheOrDB", "cache", "hit")

		case CacheEncoderTypeGob:
			// decode gob stored in cache as string
			b, err := cacheResult.AsBytes()
			if err != nil {
				slog.Warn("could not get value from cache", "cache", "error", "error", err)
				return value, err
			}

			value, err = DecodeGob[T](b)
			if err != nil {
				slog.Warn("could not decode value from cache", "cache", "error", "error", err)
				return value, err
			}

			slog.Debug("service.Cache.FromCacheOrDB", "cache", "hit")
		}

		// value found but no was possible to decode
		return value, cacheResult.Error()

		// value not found in cache or error was returned (when server is down)
	} else if cacheResult.Error() != nil {
		slog.Debug("service.Cache.FromCacheOrDB", "cache", "miss")

		// try to query the value
		value, err := query()
		if err != nil {
			slog.Warn("could not query value", "cache", "error", "error", err)
			return value, err
		}

		switch cacheType {
		case CacheEncoderTypeJSON:
			str, err := EncodeJSON(value)
			if err != nil {
				slog.Warn("could not encode value to set in cache", "cache", "error", "error", err)
				return value, nil
			}

			if len(str) == 0 {
				slog.Warn("could not encode value to set in cache", "cache", "error", "error", err)
				return value, nil
			}

			slog.Debug("service.Cache.FromCacheOrDB", "cache", "set", "key", key)

			// set in cache, if it fails, no error is returned
			cacheResult = cache.DoCache(ctx, valkey.Cacheable(cache.B().Set().Key(key).Value(str).Build()), ttl)
			if cacheResult.Error() != nil {
				slog.Warn("could not set value in cache", "cache", "error", "error", cacheResult.Error())
				return value, nil
			}

		case CacheEncoderTypeGob:
			b, err := EncodeGob(value)
			if err != nil {
				slog.Warn("could not encode value to set in cache", "cache", "error", "error", err)
				return value, nil
			}

			if len(b) == 0 {
				slog.Warn("could not encode value to set in cache", "cache", "error", "error", err)
				return value, nil
			}

			slog.Debug("service.Cache.FromCacheOrDB", "cache", "set", "key", key)

			// set in cache, if it fails, no error is returned
			cacheResult = cache.DoCache(ctx, valkey.Cacheable(cache.B().Set().Key(key).Value(valkey.BinaryString(b)).Build()), ttl)
			if cacheResult.Error() != nil {
				slog.Warn("could not set value in cache", "cache", "error", "error", cacheResult.Error())
				return value, nil
			}
		}

		return value, nil
	}

	return value, nil
}

// EncodeJSON encodes the value to JSON
func EncodeJSON[T any](value T) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// EncodeGob encodes the value to gob
func EncodeGob[T any](value T) ([]byte, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(value); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// DecodeGob decodes the value from gob
func DecodeGob[T any](in []byte) (T, error) {
	var value T
	b := bytes.NewReader(in)
	if err := gob.NewDecoder(b).Decode(&value); err != nil {
		return value, err
	}

	return value, nil
}

// Remove removes the value from cache
// ctx: context used to wait get and set operations on cache server, this must be lowed to avoid blocking
// when the cache server is slow or down
// cache: cache repository to use
// key: key to use in cache
//
// if key does not exist in cache, no error is returned
func (ref *CacheService) Remove(ctx context.Context, key string) {
	cacheResult := ref.cache.DoCache(ctx, valkey.Cacheable(ref.cache.B().Del().Key(key).Build()), 0)
	if cacheResult.Error() != nil {
		slog.Warn("could not remove value from cache", "cache", "error", "error", cacheResult.Error())
		return
	}

	slog.Debug("service.Cache.FromCacheOrDB", "cache", "del", "key", key)
}
