package bus

import "github.com/patrickmn/go-cache"

// NewCache initializes a new cache storage and assigns it to the Cache field
// of Bus.
func (b *Bus) NewCache() {
	// cleanupInterval is set to 0 to avoid spinning up the janitor
	// goroutine.
	b.Cache = cache.New(cache.NoExpiration, 0)
}

// FlushCache clears the Bus cache storage, and sets the value of Bus.Cache to
// nil.
func (b *Bus) FlushCache() {
	if b.Cache != nil {
		b.Cache.Flush()
	}

	b.Cache = nil
}
