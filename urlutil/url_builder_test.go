package urlutil

import (
	"testing"
)

func checkUrl(t *testing.T, actual string, excepteds ...string) {
	found := false
	for _, excepted := range excepteds {
		if actual == excepted {
			found = true
		}
	}
	if !found {
		t.Errorf("actual is %s, excepted is %s", actual, excepteds[0])
	}
}

func TestBuilder(t *testing.T) {
	url := NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b")
	url = NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").WithQuery("c", "1").WithQuery("c", "1").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?c=1&c=1")
	url = NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").WithQueries(map[string]string{"c": "1", "b": "d"}, "").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?b=d&c=1", "http://12.12.121.1/aa/a/b?c=1&b=d")
	url = NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").WithQueries(map[string]string{"c": "1", "b": "d"}, "@").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?@b=d&@c=1", "http://12.12.121.1/aa/a/b?@c=1&@b=d")
	url = NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?b=d&c=1", "http://12.12.121.1/aa/a/b?c=1&b=d")
	url = NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "@").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?@c=1&@b=d", "http://12.12.121.1/aa/a/b?@b=d&@c=1")
	url = NewURLBuilder("http://12.12.121.1/aa").Concat("a", "b").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "@").WithQuery("f", "f").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa/a/b?@c=1&@b=d&f=f", "http://12.12.121.1/aa/a/b?@b=d&@c=1&f=f")
	url = NewURLBuilder("http://12.12.121.1/aa?").WithAnyQueries(map[string]interface{}{"c": 1, "b": "d"}, "@").WithQuery("f", "f").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa?@c=1&@b=d&f=f", "http://12.12.121.1/aa?@b=d&@c=1&f=f")
	url = NewURLBuilder("http://12.12.121.1/aa?").WithQuery("f", "f").ToUrl()
	checkUrl(t, url, "http://12.12.121.1/aa?f=f")
}
