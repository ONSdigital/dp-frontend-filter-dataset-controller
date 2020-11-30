package handlers

import (
	"errors"
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
		FilterID:   filterID,
	}

	Convey("Given a set of mocked clients", t, func() {

		mockFilterClient := NewMockFilterClient(mockCtrl)
		mockHierarchyClient := NewMockHierarchyClient(mockCtrl)

		mockHierarchyModel := hierarchy.Model{
			Children: []hierarchy.Child{
				{
					Links: hierarchy.Links{
						Code: hierarchy.Link{
							ID: "opt1",
						},
					},
				}, {
					Links: hierarchy.Links{
						Code: hierarchy.Link{
							ID: "opt2",
						},
					},
				},
			},
		}

		// prepare request for provided url and form, then perform the call with a response writer, which is returned
		callUpdateHierarchy := func(url string, form url.Values) *httptest.ResponseRecorder {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			So(err, ShouldBeNil)
			cookie := http.Cookie{Name: dprequest.CollectionIDCookieKey, Value: mockCollectionID}
			req.AddCookie(&cookie)
			req.Header.Add(dprequest.FlorenceHeaderKey, mockUserAuthToken)
			req.Form = form

			router := mux.NewRouter()
			w := httptest.NewRecorder()
			f := NewFilter(nil, mockFilterClient, nil, mockHierarchyClient, nil, nil, mockSearchAPIAuthToken, "", "/v1", false, batchSize)
			router.Path("/filters/{filterID}/dimensions/{name}/update").HandlerFunc(f.HierarchyUpdate())
			router.Path("/filters/{filterID}/dimensions/{name}/{code}/update").HandlerFunc(f.HierarchyUpdate())
			router.ServeHTTP(w, req)
			return w
		}

		Convey("HierarchyUpdate called with a form containing new and overlapping options results in a patch with all options to add and nothing to remove", func() {
			testForm := url.Values{
				"opt3": []string{"v31", "v32", "v33"},
				"opt4": []string{"v41", "v42", "v43"},
				"opt5": []string{"v51", "v52", "v53"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt3", "opt4", "opt5"}), []string{""}, batchSize).Return(nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(hierarchy.Model{}, nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("HierarchyUpdate with code and called against a model with children, "+
			"results in options present in children and not in the request being removed via a patch operation", func() {
			testForm := url.Values{
				"opt2": []string{"v31", "v32", "v33"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt2"}), []string{"opt1"}, batchSize).Return(nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(mockHierarchyModel, nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})

		Convey("Dimension HierarchyUpdate with a form containing 'add-all' results in a single call to set the options returned by hierarchy GetRoot", func() {
			testForm := url.Values{
				"add-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt1", "opt2"})).Return(nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("Dimension code HierarchyUpdated with a form containing 'add-all' results in a single call to set the options returned by hierarchy GetChild for the provided code", func() {
			testForm := url.Values{
				"add-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().SetDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{"opt1", "opt2"})).Return(nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})

		Convey("Dimension HierarchyUpdate with a form containing 'remove-all' results in a single call to patch-remove the options returned by hierarchy GetRoot", func() {
			testForm := url.Values{
				"remove-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockHierarchyClient.EXPECT().GetRoot(ctx, testInstanceID, dimensionName).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{}), ItemsEq([]string{"opt1", "opt2"}), batchSize).Return(nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension\">Found</a>.\n\n")
		})

		Convey("Dimension code HierarchyUpdated with a form containing 'remove-all' results in a single call to patch-remove the options returned by hierarchy GetChild for the provided code", func() {
			testForm := url.Values{
				"remove-all": []string{"true"},
			}

			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filterModel, nil)
			mockHierarchyClient.EXPECT().GetChild(ctx, testInstanceID, dimensionName, mockCode).Return(mockHierarchyModel, nil)
			mockFilterClient.EXPECT().PatchDimensionValues(ctx, mockUserAuthToken, "", mockCollectionID, filterID, dimensionName,
				ItemsEq([]string{}), ItemsEq([]string{"opt1", "opt2"}), batchSize).Return(nil)

			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/%s/update", filterID, dimensionName, mockCode), testForm)

			So(w.Code, ShouldEqual, http.StatusFound)
			So(w.Body.String(), ShouldEqual, "<a href=\"/filters/12345/dimensions/myDimension/testCode\">Found</a>.\n\n")
		})

		Convey("Then if GetJobState fails, the hierarchy update is aborted and a 500 status code is returned", func() {
			errGetJobState := errors.New("error getting job state")
			mockFilterClient.EXPECT().GetJobState(ctx, mockUserAuthToken, "", "", mockCollectionID, filterID).Return(filter.Model{}, errGetJobState)
			w := callUpdateHierarchy(fmt.Sprintf("/filters/%s/dimensions/%s/update", filterID, dimensionName), nil)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

}
