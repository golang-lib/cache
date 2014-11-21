## About

This is a memcache client library for the Go programming language
(http://golang.org/).

## Installing

### Using *go get*

    $ go get github.com/gopkg/cache/memcache

After this command *memcache* is ready to use. Its source will be in:

    $GOPATH/src/github.com/gopkg/cache/memcache

## Example

```Go
import (
	"github.com/gopkg/cache/memcache"
)

func main() {
	mc := memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212")
	mc.Set(&memcache.Item{Key: "foo", Value: []byte("my value")})

	it, err := mc.Get("foo")
	...
}
```

## Full docs, see:

See https://godoc.org/github.com/gopkg/cache/memcache

Or run:

    $ godoc github.com/gopkg/cache/memcache

