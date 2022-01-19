package handlers

import (
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/hierarchy"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_isHierarchicalDimension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := gomock.Any()

	var hierarchyModel hierarchy.Model

	Convey("Should return true for hierarchical dimension", t, func() {
		mockHierarchyClient := NewMockHierarchyClient(ctrl)
		mockHierarchyClient.EXPECT().GetRoot(ctx, gomock.Eq(""), gomock.Eq("")).Return(hierarchyModel, nil)

		target := &Filter{HierarchyClient: mockHierarchyClient}

		isHierarchy, err := target.isHierarchicalDimension(nil, "", "")

		So(isHierarchy, ShouldBeTrue)
		So(err, ShouldBeNil)
	})

	Convey("Should return false for non hierarchical dimension", t, func() {
		hierarchyErr := hierarchy.NewErrInvalidHierarchyAPIResponse(200, 404, "")

		mockHierarchyClient := NewMockHierarchyClient(ctrl)
		mockHierarchyClient.EXPECT().GetRoot(ctx, gomock.Eq(""), gomock.Eq("")).Return(hierarchyModel, hierarchyErr)

		target := &Filter{HierarchyClient: mockHierarchyClient}

		isHierarchy, err := target.isHierarchicalDimension(nil, "", "")

		So(isHierarchy, ShouldBeFalse)
		So(err, ShouldBeNil)
	})

	Convey("Should return error if hierarchy response is unsuccessful and not status 404", t, func() {
		hierarchyErr := hierarchy.NewErrInvalidHierarchyAPIResponse(200, 500, "")

		mockHierarchyClient := NewMockHierarchyClient(ctrl)
		mockHierarchyClient.EXPECT().GetRoot(ctx, gomock.Eq(""), gomock.Eq("")).Return(hierarchyModel, hierarchyErr)

		target := &Filter{HierarchyClient: mockHierarchyClient}

		isHierarchy, err := target.isHierarchicalDimension(nil, "", "")

		So(isHierarchy, ShouldBeFalse)
		So(err, ShouldResemble, hierarchyErr)
	})

	Convey("Should return error if hierarchy client returns a standard error", t, func() {
		err := errors.New("borked")

		mockHierarchyClient := NewMockHierarchyClient(ctrl)
		mockHierarchyClient.EXPECT().GetRoot(ctx, gomock.Eq(""), gomock.Eq("")).Return(hierarchyModel, err)

		target := &Filter{HierarchyClient: mockHierarchyClient}

		isHierarchy, err := target.isHierarchicalDimension(nil, "", "")

		So(isHierarchy, ShouldBeFalse)
		So(err, ShouldResemble, err)
	})
}
