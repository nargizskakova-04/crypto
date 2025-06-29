package cache

type RedisAdapter struct{}

func NewRedisAdapter() *RedisAdapter {
	return &RedisAdapter{}
}
