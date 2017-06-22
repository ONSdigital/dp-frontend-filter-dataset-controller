package handlers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCMD(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test landing http cmd handler", t, func() {
		expectedReqBody := `{"type":"","uri":"","taxonomy":null,"breadcrumb":null,"serviceMessage":"","metadata":{"title":"","description":"","keywords":null,"footer":{"enabled":true,"contact":"Matt Rout","release_date":"11 November 2016","next_release":"11 November 2017","dataset_id":"MR"}},"searchDisabled":true}`

		Convey("test successful request for getting cmd landing page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/startpage", []byte(expectedReqBody)).Return([]byte(`landing-page`), nil)

			c := NewCMD(mr)

			testResponse(200, "landing-page", "/datasets/1234/editions/5678/versions/2017", c.Landing)
		})

		Convey("test error thrown when rendering landing page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/startpage", []byte(expectedReqBody)).Return(nil, errors.New("something went wrong :-("))

			c := NewCMD(mr)

			testResponse(500, "", "/datasets/1234/editions/5678/versions/2017", c.Landing)
		})
	})

	Convey("test CreateJobID handler, creates a job id and redirects", t, func() {
		c := NewCMD(nil)

		w := testResponse(301, "", "/datasets/1234/editions/5678/versions/2017/filter", c.CreateJobID)

		location := w.Header().Get("Location")
		So(location, ShouldNotBeEmpty)

		matched, err := regexp.MatchString(`^\/jobs\/\d{8}$`, location)
		So(err, ShouldBeNil)
		So(matched, ShouldBeTrue)
	})

	Convey("test middle page cmd handler", t, func() {
		expectedReqBody := `{"type":"","uri":"","taxonomy":null,"breadcrumb":null,"serviceMessage":"","metadata":{"title":"","description":"","keywords":null,"footer":{"enabled":true,"contact":"Matt Rout","release_date":"11 November 2016","next_release":"11 November 2017","dataset_id":"MR"}},"searchDisabled":true,"data":{"job_id":""}}`
		Convey("test successful request for getting cmd middle page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/middlepage", []byte(expectedReqBody)).Return([]byte(`middle-page`), nil)

			c := NewCMD(mr)

			testResponse(200, "middle-page", "/jobs/12345678", c.Middle)
		})

		Convey("test error thrown when rendering middle page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/middlepage", []byte(expectedReqBody)).Return(nil, errors.New("something went wrong with middle page rendering :-("))

			c := NewCMD(mr)

			testResponse(500, "", "/jobs/12345678", c.Middle)
		})
	})

	Convey("test landing http finish handler", t, func() {
		expectedReqBody := `{"type":"","uri":"","taxonomy":null,"breadcrumb":null,"serviceMessage":"","metadata":{"title":"","description":"","keywords":null,"footer":{"enabled":true,"contact":"Matt Rout","release_date":"11 November 2016","next_release":"11 November 2017","dataset_id":"MR"}},"searchDisabled":true}`
		Convey("test successful request for getting cmd finish page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/finishpage", []byte(expectedReqBody)).Return([]byte(`finish-page`), nil)

			c := NewCMD(mr)

			testResponse(200, "finish-page", "/jobs/12345678/dimensions", c.PreviewAndDownload)
		})

		Convey("test error thrown when rendering finish page", func() {
			mr := renderer.NewMockRenderer(mockCtrl)
			mr.EXPECT().Do("dataset/finishpage", []byte(expectedReqBody)).Return(nil, errors.New("something went wrong rendering finish page :-("))

			c := NewCMD(mr)

			testResponse(500, "", "/jobs/12345678/dimensions", c.PreviewAndDownload)
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
