package collect

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/XmchxUp/goBiliBili/extensions"
	"github.com/XmchxUp/goBiliBili/logger"
	"github.com/XmchxUp/goBiliBili/proxy"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type Fetcher interface {
	Get(req *Request) ([]byte, error)
}

type Request struct {
	URL    string
	Cookie string
}

type BaseFetch struct {
}

func (BaseFetch) Get(req *Request) ([]byte, error) {
	resp, err := http.Get(req.URL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status code:%d", resp.StatusCode)
	}

	bodyReader := bufio.NewReader(resp.Body)
	e := DetermineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
	return io.ReadAll(utf8Reader)
}

// BrowserFetch 模拟浏览器访问
type BrowserFetch struct {
	Timeout           time.Duration
	Proxy             proxy.Func
	Log               *logger.Logger
	AutoConvertToUTF8 bool
}

func (b *BrowserFetch) Get(request *Request) ([]byte, error) {
	client := &http.Client{Timeout: b.Timeout}

	if b.Proxy != nil {
		transport := http.DefaultTransport.(*http.Transport)
		transport.Proxy = b.Proxy
		client.Transport = transport
	}

	req, err := http.NewRequest(http.MethodGet, request.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("get url failed:%w", err)
	}

	req.Header.Set("User-Agent", extensions.GenerateRandomUA())

	if len(request.Cookie) > 0 {
		req.Header.Set("Cookie", request.Cookie)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !b.AutoConvertToUTF8 {
		return io.ReadAll(resp.Body)
	}

	bodyReader := bufio.NewReader(resp.Body)

	e := DetermineEncoding(bodyReader)

	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return io.ReadAll(utf8Reader)
}

func DetermineEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)

	if err != nil {
		return unicode.UTF8
	}

	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}
