package handlers

import (
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateTime(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()
	options := gomock.Any()
	const mockUserAuthToken = "Foo"
	const mockServiceAuthToken = ""
	const mockCollectionID = "Bar"
	const mockFilterID = ""
	Convey("test update time function", t, func() {
		Convey("given a valid list of options there should be no errors but be a redirect", func() {
			mockClient := NewMockFilterClient(mockCtrl)
			mockClient.EXPECT().RemoveDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "time").Return(nil)
			mockClient.EXPECT().AddDimension(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "time").Return(nil)
			mockClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "time").Return([]filter.DimensionOption{}, nil)
			mockClient.EXPECT().AddDimensionValues(ctx, mockUserAuthToken, mockServiceAuthToken, mockCollectionID, mockFilterID, "time", options).Return(nil) // Might not be able to use gomock.Any() here
			target := fmt.Sprintf("/filters/%s/dimensions/time/update", mockFilterID)
			formData := "latest-option=Nov-17&latest-month=November&latest-year=2017&month-single=Select&year-single=Select&start-month=Select&start-year=Select&end-month=Select&end-year=Select&time-selection=list&August=August&start-year-grouped=2011&end-year-grouped=2012&save-and-return=Save+and+return"
			reader := strings.NewReader(formData)
			req := httptest.NewRequest("POST", target, reader)
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Header.Add(dprequest.CollectionIDHeaderKey, mockCollectionID)
			w := httptest.NewRecorder()
			f := NewFilter(nil, mockClient, nil, nil, nil, nil, mockServiceAuthToken, "", "/v1", false)
			f.UpdateTime().ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 302)

		})
	})
}
