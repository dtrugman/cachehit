package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultExpiration = 0 * time.Second
)

type options struct {
	expiration time.Duration
}

func defaultOptions() *options {
	return &options{
		expiration: DefaultExpiration,
	}
}

func compileOptions(opts ...Option) *options {
	o := defaultOptions()

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type Option func(*options)

func WithExpiration(expiration time.Duration) Option {
	return func(o *options) {
		o.expiration = expiration
	}
}

type Redis[K comparable, V any] struct {
	underlying *redis.Client
	expiration time.Duration
}

func From[K comparable, V any](
	underlying *redis.Client,
	opts ...Option,
) *Redis[K, V] {
	o := compileOptions(opts...)
	return &Redis[K, V]{
		underlying: underlying,
		expiration: o.expiration,
	}
}

func (r *Redis[K, V]) Get(ctx context.Context, key K) (V, bool) {
	var zero V

	keyStr := fmt.Sprintf("%v", key)

	cmd := r.underlying.Get(ctx, keyStr)
	if cmd.Err() != nil {
		return zero, false
	}
	rawValue := cmd.Val()

	var value V
	var err error

	switch ptr := any(&value).(type) {
	case *string:
		*ptr = rawValue
	case *bool:
		parsed, parseErr := strconv.ParseBool(rawValue)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = parsed
		}
	case *int:
		parsed, parseErr := strconv.Atoi(rawValue)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = parsed
		}
	case *int8:
		parsed, parseErr := strconv.ParseInt(rawValue, 10, 8)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = int8(parsed)
		}
	case *int16:
		parsed, parseErr := strconv.ParseInt(rawValue, 10, 16)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = int16(parsed)
		}
	case *int32:
		parsed, parseErr := strconv.ParseInt(rawValue, 10, 32)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = int32(parsed)
		}
	case *int64:
		parsed, parseErr := strconv.ParseInt(rawValue, 10, 64)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = parsed
		}
	case *uint:
		parsed, parseErr := strconv.ParseUint(rawValue, 10, 0)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = uint(parsed)
		}
	case *uint8:
		parsed, parseErr := strconv.ParseUint(rawValue, 10, 8)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = uint8(parsed)
		}
	case *uint16:
		parsed, parseErr := strconv.ParseUint(rawValue, 10, 16)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = uint16(parsed)
		}
	case *uint32:
		parsed, parseErr := strconv.ParseUint(rawValue, 10, 32)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = uint32(parsed)
		}
	case *uint64:
		parsed, parseErr := strconv.ParseUint(rawValue, 10, 64)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = parsed
		}
	case *float32:
		parsed, parseErr := strconv.ParseFloat(rawValue, 32)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = float32(parsed)
		}
	case *float64:
		parsed, parseErr := strconv.ParseFloat(rawValue, 64)
		if parseErr != nil {
			err = parseErr
		} else {
			*ptr = parsed
		}
	default:
		err = json.Unmarshal([]byte(rawValue), &value)
	}

	if err != nil {
		return zero, false
	}

	return value, true
}

func (r *Redis[K, V]) Set(ctx context.Context, key K, value V) {
	keyStr := fmt.Sprintf("%v", key)

	var valueStr string

	switch v := any(value).(type) {
	case string:
		valueStr = v
	case bool:
		valueStr = strconv.FormatBool(v)
	case int:
		valueStr = strconv.Itoa(v)
	case int8:
		valueStr = strconv.FormatInt(int64(v), 10)
	case int16:
		valueStr = strconv.FormatInt(int64(v), 10)
	case int32:
		valueStr = strconv.FormatInt(int64(v), 10)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	case uint:
		valueStr = strconv.FormatUint(uint64(v), 10)
	case uint8:
		valueStr = strconv.FormatUint(uint64(v), 10)
	case uint16:
		valueStr = strconv.FormatUint(uint64(v), 10)
	case uint32:
		valueStr = strconv.FormatUint(uint64(v), 10)
	case uint64:
		valueStr = strconv.FormatUint(v, 10)
	case float32:
		valueStr = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return
		}
		valueStr = string(valueBytes)
	}

	_ = r.underlying.Set(ctx, keyStr, valueStr, r.expiration)
}
