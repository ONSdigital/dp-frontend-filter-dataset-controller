package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	dprequest "github.com/ONSdigital/dp-net/request"
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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ctx := gomock.Any()

	mockSearchAPIAuthToken := "testServiceAuthToken"
	mockUserAuthToken := "testUserAuthToken"
	mockCollectionID := "testCollectionID"

	filterID := "12345"
	dimensionName := "myDimension"
	mockCode := "testCode"
	testInstanceID := "testInstanceID"
	batchSize := 100
	filterModel := filter.Model{
		InstanceID: testInstanceID,
	}

	Convey("Given that filter API has three options for the dimension under test", t, func() {
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
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt3", "opt4", "opt5"}), []string{""}, batchSize).Return(nil)

			// HierarchyClient mock expecting GetRoot
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)

			// Prepare request with header
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), nil)
			cookie := http.Cookie{Name: dprequest.CollectionIDCookieKey, Value: mockCollectionID}
			req.AddCookie(&cookie)
			So(err, ShouldBeNil)
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Form = testForm

			w := httptest.NewRecorder()

			// Set handler and perform call
			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", "/v1", false, batchSize)
			router.Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(f.HierarchyUpdate())
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("HierarchyUpdate with code and called against a model with children, "+
			"results in only options present in children and not in the request being removed from the filter API options", func() {

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

			modelWithChildren := hierarchy.Model{Children: []hierarchy.Child{child1, child2}}

			// We call FilterAPI AddDimensionValue for the option provided in the form, and RemoveDimensionValue for the opt1 because it is present in children but not in request form
			mockFilterClient := NewMockFilterClient(mockCtrl)
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt2"}), []string{"opt1"}, batchSize).Return(nil)

			// HierarchyClient mock expecting GetChild, returns child model with self link
			mockHierarchyClient := NewMockHierarchyClient(mockCtrl)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(modelWithChildren, nil)

			// Prepare request with header and context
			req, err := http.NewRequest("GET", fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), nil)
			cookie := http.Cookie{Name: dprequest.CollectionIDCookieKey, Value: mockCollectionID}
			req.AddCookie(&cookie)
			So(err, ShouldBeNil)
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Form = testForm

			w := httptest.NewRecorder()

			// Set handler and perform call
			router := mux.NewRouter()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", "/v1", false, batchSize)
			router.Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(f.HierarchyUpdate())
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})
	})

}
