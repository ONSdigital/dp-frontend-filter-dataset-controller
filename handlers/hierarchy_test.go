package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

// Header keys constants
const (
	CollectionIDHeaderKey  = "Collection-Id"
	FlorenceTokenHeaderKey = "X-Florence-Token"
)

func TestHierarchyUpdate(t *testing.T) {
	ctx := gomock.Any()

	mockSearchAPIAuthToken := "testServiceAuthToken"
	mockUserAuthToken := "testUserAuthToken"
	mockCollectionID := "testCollectionID"
	filterID := "12345"
	dimensionName := "myDimension"
	testInstanceID := "testInstanceID"
	ctxWithCollectionID := context.WithValue(context.Background(), CollectionIDHeaderKey, mockCollectionID)

	filterModel := filter.Model{
		InstanceID: testInstanceID,
	}

	dimensionOptions := []filter.DimensionOption{
		filter.DimensionOption{
			Option:              "dimensionOption1",
			DimensionOptionsURL: "http://dimension.option.1.co.uk",
		},
		filter.DimensionOption{
			Option:              "dimensionOption2",
			DimensionOptionsURL: "http://dimension.option.2.co.uk",
		},
		filter.DimensionOption{
			Option:              "dimensionOption3",
			DimensionOptionsURL: "http://dimension.option.3.co.uk",
		},
	}

	// Provide a larger number than the batch size
	testForm := url.Values{
		"k1": []string{"v11", "v12", "v13"},
		"k2": []string{"v21", "v22", "v23"},
		"k3": []string{"v31", "v32", "v33"},
		"k4": []string{"v41", "v42", "v43"},
		"k5": []string{"v51", "v52", "v53"},
	}
	batchSizeFilterAPI := 3

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test HierarchyUpdate", t, func() {
		Convey("test HierarchyUpdate", func() {

			// FilterClient mock expecting 2 AddDimensionValues batch calls - one with 3 items and one with 2 items
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName).Return(dimensionOptions, nil)
			mockFilterClient.EXPECT().AddDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, gomock.Len(3)).Return(nil)
			mockFilterClient.EXPECT().AddDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, gomock.Len(2)).Return(nil)

			// HierarchyClient mock expecting GetRoot
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)

			// Perform the call
			req, err := http.NewRequestWithContext(ctxWithCollectionID, "GET", fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), nil)
			So(err, ShouldBeNil)
			req.Header.Add("X-Florence-Token", mockUserAuthToken)
			req.Form = testForm

			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", false, batchSizeFilterAPI)
			router.Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(f.HierarchyUpdate)
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("test HierarchyUpdate with code", func() {

			// test code
			mockCode := "testCode"

			// hierarchy child, with grandchildren that match a dimensionOption
			child := hierarchy.Model{
				Children: []hierarchy.Child{
					hierarchy.Child{
						Links: hierarchy.Links{
							Self: hierarchy.Link{
								ID: "dimensionOption1", // Present in dimensionOptions
							},
						},
					},
				},
			}

			// FilterClient mock expecting 2 AddDimensionValues batch calls - one with 3 items and one with 2 items
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName).Return(dimensionOptions, nil)
			mockFilterClient.EXPECT().AddDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, gomock.Len(3)).Return(nil)
			mockFilterClient.EXPECT().AddDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, gomock.Len(2)).Return(nil)
			mockFilterClient.EXPECT().RemoveDimensionValue(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, "dimensionOption1").Return(nil)

			// HierarchyClient mock expecting GetChild, returns child model with self link
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(child, nil)

			// Perform the call
			req, err := http.NewRequestWithContext(ctxWithCollectionID, "GET", fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), nil)
			So(err, ShouldBeNil)
			req.Header.Add("X-Florence-Token", mockUserAuthToken)
			req.Form = testForm

			w := httptest.NewRecorder()

			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", false, batchSizeFilterAPI)
			router.Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(f.HierarchyUpdate)
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})
	})

}
