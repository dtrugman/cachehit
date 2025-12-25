package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/dtrugman/cachehit/example/resource"
)

func TestFrom(t *testing.T) {
	client := redis.NewClient(&redis.Options{})

	adapter := From[string, int](client)
	require.NotNil(t, adapter)
	require.Equal(t, client, adapter.underlying)
	require.Equal(t, DefaultExpiration, adapter.expiration)
}

func TestFrom_WithExpiration(t *testing.T) {
	client := redis.NewClient(&redis.Options{})

	expiration := 5 * time.Second
	adapter := From[string, int](client, WithExpiration(expiration))
	require.NotNil(t, adapter)
	require.Equal(t, expiration, adapter.expiration)
}

func TestRedis_Error(t *testing.T) {
	ctx := context.Background()

	disconnectedClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999",
	})
	adapter := From[string, string](disconnectedClient)

	key := uuid.New().String()
	_, err := adapter.Get(ctx, key)
	require.Error(t, err)
}

func TestRedis_Operations(t *testing.T) {
	ctx := context.Background()

	instance, err := resource.RedisRun(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, instance.Cleanup())
	})

	client, err := resource.RedisConn(ctx, instance.DSN)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
	})

	t.Run("String", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, string](client)

		adapter.Set(ctx, key, "value1")

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, "value1", value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "value1", redisValue)

		adapter.Set(ctx, key, "value2")

		value, err = adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, "value2", value)

		redisValue, err = client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "value2", redisValue)

		value, err = adapter.Get(ctx, uuid.New().String())
		require.Error(t, err)
		require.Equal(t, "", value)
	})

	t.Run("Int", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, int](client)

		adapter.Set(ctx, key, 42)

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, 42, value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "42", redisValue)
	})

	t.Run("Int8", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, int8](client)

		adapter.Set(ctx, key, int8(127))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, int8(127), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "127", redisValue)
	})

	t.Run("Int16", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, int16](client)

		adapter.Set(ctx, key, int16(32767))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, int16(32767), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "32767", redisValue)
	})

	t.Run("Int32", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, int32](client)

		adapter.Set(ctx, key, int32(2147483647))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, int32(2147483647), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "2147483647", redisValue)
	})

	t.Run("Int64", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, int64](client)

		adapter.Set(ctx, key, int64(9223372036854775807))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, int64(9223372036854775807), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "9223372036854775807", redisValue)
	})

	t.Run("Uint", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, uint](client)

		adapter.Set(ctx, key, uint(42))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, uint(42), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "42", redisValue)
	})

	t.Run("Uint8", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, uint8](client)

		adapter.Set(ctx, key, uint8(255))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, uint8(255), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "255", redisValue)
	})

	t.Run("Uint16", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, uint16](client)

		adapter.Set(ctx, key, uint16(65535))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, uint16(65535), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "65535", redisValue)
	})

	t.Run("Uint32", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, uint32](client)

		adapter.Set(ctx, key, uint32(4294967295))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, uint32(4294967295), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "4294967295", redisValue)
	})

	t.Run("Uint64", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, uint64](client)

		adapter.Set(ctx, key, uint64(18446744073709551615))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, uint64(18446744073709551615), value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "18446744073709551615", redisValue)
	})

	t.Run("Bool", func(t *testing.T) {
		key1 := uuid.New().String()
		key2 := uuid.New().String()
		adapter := From[string, bool](client)

		adapter.Set(ctx, key1, true)

		value, err := adapter.Get(ctx, key1)
		require.NoError(t, err)
		require.Equal(t, true, value)

		redisValue, err := client.Get(ctx, key1).Result()
		require.NoError(t, err)
		require.Equal(t, "true", redisValue)

		adapter.Set(ctx, key2, false)

		value, err = adapter.Get(ctx, key2)
		require.NoError(t, err)
		require.Equal(t, false, value)

		redisValue, err = client.Get(ctx, key2).Result()
		require.NoError(t, err)
		require.Equal(t, "false", redisValue)
	})

	t.Run("Float32", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, float32](client)

		adapter.Set(ctx, key, float32(3.14))

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.InDelta(t, float32(3.14), value, 0.01)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "3.14", redisValue)
	})

	t.Run("Float64", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, float64](client)

		adapter.Set(ctx, key, 3.14159)

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.InDelta(t, 3.14159, value, 0.01)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, "3.14159", redisValue)
	})

	t.Run("Struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}

		key := uuid.New().String()
		adapter := From[string, TestStruct](client)

		testData := TestStruct{Name: "test", Value: 123}
		adapter.Set(ctx, key, testData)

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, testData, value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, `{"Name":"test","Value":123}`, redisValue)
	})

	t.Run("Array", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, []int](client)

		testData := []int{1, 2, 3, 4, 5}
		adapter.Set(ctx, key, testData)

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, testData, value)

		redisValue, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		require.Equal(t, `[1,2,3,4,5]`, redisValue)
	})

	t.Run("Expiration", func(t *testing.T) {
		key := uuid.New().String()
		adapter := From[string, string](client, WithExpiration(1*time.Second))

		adapter.Set(ctx, key, "value1")

		value, err := adapter.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, "value1", value)

		time.Sleep(2 * time.Second)

		_, err = adapter.Get(ctx, key)
		require.Error(t, err)
	})

	t.Run("IntKey", func(t *testing.T) {
		adapter := From[int, string](client)

		intKey := 123456789
		adapter.Set(ctx, intKey, "value1")

		value, err := adapter.Get(ctx, intKey)
		require.NoError(t, err)
		require.Equal(t, "value1", value)
	})

	t.Run("ParsingErrors", func(t *testing.T) {
		t.Run("Bool", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, bool](client)

			cmd := client.Set(ctx, key, "invalid", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Int", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, int](client)

			cmd := client.Set(ctx, key, "not-a-number", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Int8", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, int8](client)

			cmd := client.Set(ctx, key, "999", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Int16", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, int16](client)

			cmd := client.Set(ctx, key, "99999", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Int32", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, int32](client)

			cmd := client.Set(ctx, key, "9999999999", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Int64", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, int64](client)

			cmd := client.Set(ctx, key, "not-a-number", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Uint", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, uint](client)

			cmd := client.Set(ctx, key, "-42", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Uint8", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, uint8](client)

			cmd := client.Set(ctx, key, "999", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Uint16", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, uint16](client)

			cmd := client.Set(ctx, key, "99999", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Uint32", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, uint32](client)

			cmd := client.Set(ctx, key, "9999999999", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Uint64", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, uint64](client)

			cmd := client.Set(ctx, key, "-1", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Float32", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, float32](client)

			cmd := client.Set(ctx, key, "not-a-float", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("Float64", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, float64](client)

			cmd := client.Set(ctx, key, "not-a-float", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})

		t.Run("JSON", func(t *testing.T) {
			type TestStruct struct {
				Name  string
				Value int
			}

			key := uuid.New().String()
			adapter := From[string, TestStruct](client)

			cmd := client.Set(ctx, key, "invalid-json{", 0)
			require.NoError(t, cmd.Err())

			_, err := adapter.Get(ctx, key)
			require.Error(t, err)
		})
	})

	t.Run("MarshallingErrors", func(t *testing.T) {
		t.Run("UnsupportedType", func(t *testing.T) {
			key := uuid.New().String()
			adapter := From[string, chan int](client)

			err := adapter.Set(ctx, key, make(chan int))
			require.Error(t, err)
		})

		t.Run("UnsupportedTypeInStruct", func(t *testing.T) {
			type InvalidStruct struct {
				Ch chan int
			}

			key := uuid.New().String()
			adapter := From[string, InvalidStruct](client)

			err := adapter.Set(ctx, key, InvalidStruct{Ch: make(chan int)})
			require.Error(t, err)
		})
	})
}
