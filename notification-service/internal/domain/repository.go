package domain

type CacheRepository interface {
	Save(key string, value interface{}) error
	Get(key string) *Post
}
