package renderer

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func newMockClient(status int, body []byte) *http.Client {
	client := http.DefaultClient

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(status)
		w.Write(body)
	}))

	client.Transport = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(s.URL)
		},
	}

	return client
}

func TestUnitRenderer(t *testing.T) {
	Convey("test successful response received from renderer service", t, func() {
		testHTML := `<html><h1>Hello World</h1></html>`

		c := newMockClient(200, []byte(testHTML))
		r := &renderer{client: c, url: "http://localhost:28282"}

		b, err := r.Do("/helloworld", nil)
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, testHTML)
	})

	Convey("test error returned if client url is missing", t, func() {
		os.Setenv("RENDERER_URL", "incorrect-url")
		r := New()

		b, err := r.Do("/helloworld", nil)
		So(err, ShouldNotBeNil)
		So(b, ShouldBeNil)
	})

	Convey("test error returned if invalid status code returned from renderer", t, func() {
		c := newMockClient(408, nil)

		r := &renderer{client: c, url: "http://localhost:28263"}

		b, err := r.Do("/helloworld", nil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid response from renderer service - status 408")
		So(b, ShouldBeNil)
	})
}
