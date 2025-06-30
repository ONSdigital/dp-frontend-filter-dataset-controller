package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	dpApiClientsGoDataset "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-router/router/routertest"
)

func NewHandlerMock() *routertest.HandlerMock {
	return &routertest.HandlerMock{
		ServeHTTPFunc: func(in1 http.ResponseWriter, in2 *http.Request) {},
	}
}

func TestFilterPageHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockContext := gomock.Any()
	mockFilterClient := NewMockFilterClient(mockCtrl)
	mockDatasetClient := NewMockDatasetAPISdkClient(mockCtrl)

	mockFilterModel := &filter.Model{
		Dataset: filter.Dataset{
			DatasetID: "123",
		},
	}

	collectionID := ""
	downloadServiceToken := ""
	serviceAuthToken := ""
	userAuthToken := ""
	
	headers := dpDatasetApiSdk.Headers{
		CollectionID:         collectionID,
		DownloadServiceToken: downloadServiceToken,
		ServiceToken:         serviceAuthToken,
		UserAccessToken:      userAuthToken,
	}

	Convey("Given a FilterPageHandler", t, func() {
		Convey("When filterID is missing", func() {
			filterHandler := NewHandlerMock()
			filterFlexHandler := NewHandlerMock()
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/missing-filter-id/dimensions", http.NoBody)

			// Override behaviour — simulate the missing filter id
			vars := map[string]string{"filterID": ""}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			handler := FilterPageHandler(mockFilterClient, mockDatasetClient, filterHandler, filterFlexHandler)
			handler(mockRequestWriter, mockRequest)

			Convey("Then the status code is 400", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusBadRequest)
			})
		})

		Convey("When GetJobState returns error", func() {
			filterHandler := NewHandlerMock()
			filterFlexHandler := NewHandlerMock()

			handler := FilterPageHandler(mockFilterClient, mockDatasetClient, filterHandler, filterFlexHandler)
			router := mux.NewRouter()
			router.HandleFunc("/filters/{filterID}/dimensions", handler)
			router.HandleFunc("/filters/{filterID}/dimensions/{dimension}", handler)

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions", http.NoBody)

			mockFilterClient.EXPECT().GetJobState(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, "123",
			).Return(filter.Model{}, "", errors.New("some error"))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 500", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When error is returned fetching dataset", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions", http.NoBody)

			filterHandler := NewHandlerMock()
			filterFlexHandler := NewHandlerMock()

			router := mux.NewRouter()
			handler := FilterPageHandler(mockFilterClient, mockDatasetClient, filterHandler, filterFlexHandler)
			router.HandleFunc("/filters/{filterID}/dimensions", handler)
			router.HandleFunc("/filters/{filterID}/dimensions/{dimension}", handler)

			mockFilterClient.EXPECT().GetJobState(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, "123",
			).Return(*mockFilterModel, "", nil)

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, "123",
			).Return(dpApiClientsGoDataset.DatasetDetails{}, errors.New("dataset error"))

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 500", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When dataset is 'cantabular' type", func() {
			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions/ltla", http.NoBody)

			filterHandler := NewHandlerMock()
			filterFlexHandler := NewHandlerMock()

			router := mux.NewRouter()
			handler := FilterPageHandler(mockFilterClient, mockDatasetClient, filterHandler, filterFlexHandler)
			router.HandleFunc("/filters/{filterID}/dimensions", handler)
			router.HandleFunc("/filters/{filterID}/dimensions/{dimension}", handler)

			datasetDetails := dpApiClientsGoDataset.DatasetDetails{
				Type: "cantabular",
			}

			mockFilterClient.EXPECT().GetJobState(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, "123",
			).Return(*mockFilterModel, "", nil)

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, "123",
			).Return(datasetDetails, nil)

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 200", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the request is sent to Filter Flex Dataset Service", func() {
				So(len(filterFlexHandler.ServeHTTPCalls()), ShouldEqual, 1)
				So(filterFlexHandler.ServeHTTPCalls()[0].In2.URL.Path, ShouldResemble, "/filters/123/dimensions/ltla")
			})

			Convey("Then Filter Dataset Service was not called", func() {
				So(len(filterHandler.ServeHTTPCalls()), ShouldEqual, 0)
			})
		})

		Convey("When dataset is not 'cantabular' type", func() {
			filterHandler := NewHandlerMock()
			filterFlexHandler := NewHandlerMock()

			mockRequestWriter := httptest.NewRecorder()
			mockRequest := httptest.NewRequest(http.MethodGet, "/filters/123/dimensions", http.NoBody)

			router := mux.NewRouter()
			handler := FilterPageHandler(mockFilterClient, mockDatasetClient, filterHandler, filterFlexHandler)
			router.HandleFunc("/filters/{filterID}/dimensions", handler)
			router.HandleFunc("/filters/{filterID}/dimensions/{dimension}", handler)

			datasetDetails := dpApiClientsGoDataset.DatasetDetails{
				Type: "cmd",
			}

			mockFilterClient.EXPECT().GetJobState(
				mockContext, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, "123",
			).Return(*mockFilterModel, "", nil)

			mockDatasetClient.EXPECT().GetDataset(
				mockContext, headers, collectionID, "123",
			).Return(datasetDetails, nil)

			router.ServeHTTP(mockRequestWriter, mockRequest)

			Convey("Then the status code is 200", func() {
				So(mockRequestWriter.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the request is sent to Filter Flex Dataset Service", func() {
				So(len(filterHandler.ServeHTTPCalls()), ShouldEqual, 1)
				So(filterHandler.ServeHTTPCalls()[0].In2.URL.Path, ShouldResemble, "/filters/123/dimensions")
			})

			Convey("Then Filter Dataset Service was not called", func() {
				So(len(filterFlexHandler.ServeHTTPCalls()), ShouldEqual, 0)
			})
		})
	})
}
