package handlers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/ONSdigital/dp-frontend-dataset-controller/renderer"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCMD(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test landing http cmd handler", t, func() {
		Convey("test successful request for getting cmd landing page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/startpage", nil).Return([]byte(`landing-page`), nil)

			c := NewCMD(mr)

			testResponse(200, "landing-page", "/dataset/cmd", c.Landing)
		})

		Convey("test error thrown when rendering landing page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/startpage", nil).Return(nil, errors.New("something went wrong :-("))

			c := NewCMD(mr)

			testResponse(500, "", "/dataset/cmd", c.Landing)
		})
	})

	Convey("test CreateJobID handler, creates a job id and redirects", t, func() {
		c := NewCMD(nil)

		w := testResponse(301, "", "/dataset/cmd/middle", c.CreateJobID)

		location := w.Header().Get("Location")
		So(location, ShouldNotBeEmpty)

		matched, err := regexp.MatchString(`^\/dataset\/cmd\/\d{8}$`, location)
		So(err, ShouldBeNil)
		So(matched, ShouldBeTrue)
	})

	Convey("test middle page cmd handler", t, func() {
		Convey("test successful request for getting cmd middle page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/middlepage", []byte(`{"data":{"job_id":""}}`)).Return([]byte(`middle-page`), nil)

			c := NewCMD(mr)

			testResponse(200, "middle-page", "/dataset/cmd/12345678", c.Middle)
		})

		Convey("test error thrown when rendering middle page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/middlepage", []byte(`{"data":{"job_id":""}}`)).Return(nil, errors.New("something went wrong with middle page rendering :-("))

			c := NewCMD(mr)

			testResponse(500, "", "/dataset/cmd/12345678", c.Middle)
		})
	})

	Convey("test landing http finish handler", t, func() {
		Convey("test successful request for getting cmd finish page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/finishpage", nil).Return([]byte(`finish-page`), nil)

			c := NewCMD(mr)

			testResponse(200, "finish-page", "/dataset/cmd/12345678/finish", c.PreviewAndDownload)
		})

		Convey("test error thrown when rendering finish page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/finishpage", nil).Return(nil, errors.New("something went wrong rendering finish page :-("))

			c := NewCMD(mr)

			testResponse(500, "", "/dataset/cmd/12345678/finish", c.PreviewAndDownload)
		})
	})

}

func testResponse(code int, respBody, url string, f http.HandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", url, nil)
	So(err, ShouldBeNil)

	w := httptest.NewRecorder()
	f(w, req)

	So(w.Code, ShouldEqual, code)

	b, err := ioutil.ReadAll(w.Body)
	So(err, ShouldBeNil)

	So(string(b), ShouldEqual, respBody)

	return w
}
