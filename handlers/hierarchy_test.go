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
	mockCode := "testCode"
	testInstanceID := "testInstanceID"
	ctxWithCollectionID := context.WithValue(context.Background(), CollectionIDHeaderKey, mockCollectionID)
	filterModel := filter.Model{
		InstanceID: testInstanceID,
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("Given that filter API has three options for the dimension under test", t, func() {

		// dimension options originally existing in filter API before the test
		dimensionOptions := []filter.DimensionOption{
			{
				Option:              "opt1",
				DimensionOptionsURL: "http://dimension.opt1.1.co.uk",
			},
			{
				Option:              "opt2",
				DimensionOptionsURL: "http://dimension.opt1.2.co.uk",
			},
			{
				Option:              "opt3",
				DimensionOptionsURL: "http://dimension.opt1.3.co.uk",
			},
		}

		Convey("HierarchyUpdate called with a form containing new and overlapping options results in the union of options being sent tot filter API, one by one", func() {

			// Options comming from the request form
			testForm := url.Values{
				"opt3": []string{"v31", "v32", "v33"},
				"opt4": []string{"v41", "v42", "v43"},
				"opt5": []string{"v51", "v52", "v53"},
			}

			// We call FilterAPI AddDimensionValue for each provided option in the form
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName).Return(dimensionOptions, nil)
			mockFilterClient.EXPECT().AddDimensionValue(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, "opt3").Return(nil)
			mockFilterClient.EXPECT().AddDimensionValue(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, "opt4").Return(nil)
			mockFilterClient.EXPECT().AddDimensionValue(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, "opt5").Return(nil)

			// HierarchyClient mock expecting GetRoot
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)

			// Prepare request with header
			req, err := http.NewRequestWithContext(ctxWithCollectionID, http.MethodGet, fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), nil)
			So(err, ShouldBeNil)
			req.Header.Add("X-Florence-Token", mockUserAuthToken)
			req.Form = testForm

			w := httptest.NewRecorder()

			// Set handler and perform call
			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", "/v1", false)
			router.Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(f.HierarchyUpdate)
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("HierarchyUpdate with code and called against a model with children, "+
			"results in only options present in children and not in the request being removed from the existing filter API options", func() {

			// Options comming from the request form
			testForm := url.Values{
				"opt2": []string{"v31", "v32", "v33"},
			}

			// dimension option present in filter API before test and NOT in form values
			child1 := hierarchy.Child{
				Links: hierarchy.Links{
					Self: hierarchy.Link{
						ID: "opt1",
					},
				},
			}

			// dimension option present in filter API before test and also in form values
			child2 := hierarchy.Child{
				Links: hierarchy.Links{
					Self: hierarchy.Link{
						ID: "opt2",
					},
				},
			}

			// dimension option NOT present in filter API before test and neither in form values
			childN := hierarchy.Child{
				Links: hierarchy.Links{
					Self: hierarchy.Link{
						ID: "optN",
					},
				},
			}

			modelWithChildren := hierarchy.Model{Children: []hierarchy.Child{child1, child2, childN}}

			// We call FilterAPI AddDimensionValue for the option provided in the form, and RemoveDimensionValue for the opt1 because it is present in children but not in request form
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().GetDimensionOptions(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName).Return(dimensionOptions, nil)
			mockFilterClient.EXPECT().RemoveDimensionValue(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, "opt1").Return(nil)
			mockFilterClient.EXPECT().AddDimensionValue(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName, "opt2").Return(nil)

			// HierarchyClient mock expecting GetChild, returns child model with self link
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(modelWithChildren, nil)

			// Prepare request with header and context
			req, err := http.NewRequestWithContext(ctxWithCollectionID, "GET", fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), nil)
			So(err, ShouldBeNil)
			req.Header.Add("X-Florence-Token", mockUserAuthToken)
			req.Form = testForm

			w := httptest.NewRecorder()

			// Set handler and perform call
			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", "/v1", false)
			router.Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(f.HierarchyUpdate)
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})
	})

}

// go-mock tailored matcher to compare lists of strings ignoring order
type itemsEq struct{ expected []string }

func ItemsEq(expected []string) gomock.Matcher {
	return &itemsEq{expected}
}

func (i *itemsEq) Matches(x interface{}) bool {
	mExpected := make(map[string]struct{})
	for _, e := range i.expected {
		mExpected[e] = struct{}{}
	}
	for _, val := range x.([]string) {
		if _, found := mExpected[val]; !found {
			return false
		}
	}
	return true
}

func (i *itemsEq) String() string {
	return fmt.Sprintf("%v (in any order)", i.expected)
}
