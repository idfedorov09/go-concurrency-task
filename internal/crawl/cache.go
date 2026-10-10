package crawl

import (
	"sync"
	sf "golang.org/x/sync/singleflight"
)

type cachedPageInfo struct {
	PageInfo
	err error
}

type Cache struct {
	m map[string] cachedPageInfo
	mu sync.RWMutex
	g sf.Group
}

func MakeCache() Cache {
	return Cache{
		m: make(map[string]cachedPageInfo),
	}
}

func (mc *Cache) Has(targetUrl string) (PageInfo, error, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	if v, ok := mc.m[targetUrl]; ok {
		return v.PageInfo, v.err, true
	}
	return PageInfo{}, nil, false
}

func (mc *Cache) PutResult(targetUrl string, handler FetchFunc) (PageInfo, error, bool) {
	v, err, shared := mc.g.Do(targetUrl, func() (any, error) {
		if p, err, ok := mc.Has(targetUrl); ok {
			return p, err
		}
		p, err := handler(targetUrl)
		res := cachedPageInfo{p, err}

		mc.mu.Lock()
		defer mc.mu.Unlock()

		mc.m[targetUrl] = res

		return res.PageInfo, res.err
	})
	p := v.(PageInfo)
	return p, err, shared
}
