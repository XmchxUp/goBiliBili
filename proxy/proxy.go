package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
)

type Func func(*http.Request) (*url.URL, error)

// RoundRobinProxySwitcher creates a proxy switcher function which rotates
// ProxyURLs on every request.
// The proxy type is determined by the URL scheme. "http", "https"
// and "socks5" are supported. If the scheme is empty,
// "http" is assumed.
func RoundRobinProxySwitcher(ProxyURLs ...string) (Func, error) {
	if len(ProxyURLs) < 1 {
		return nil, errors.New("proxy URL list is empty")
	}

	urls := make([]*url.URL, len(ProxyURLs))

	for i, u := range ProxyURLs {
		parsedU, err := url.Parse(u)

		if err != nil {
			return nil, err
		}

		urls[i] = parsedU
	}

	return (&roundRobinSwitcher{urls, 0}).GetProxy, nil
}

type roundRobinSwitcher struct {
	proxyURLs []*url.URL
	index     uint32
}

func (r *roundRobinSwitcher) GetProxy(_ *http.Request) (*url.URL, error) {
	if len(r.proxyURLs) == 0 {
		return nil, errors.New("empty proxy urls")
	}

	index := atomic.AddUint32(&r.index, 1) - 1
	u := r.proxyURLs[index%uint32(len(r.proxyURLs))]

	return u, nil
}
