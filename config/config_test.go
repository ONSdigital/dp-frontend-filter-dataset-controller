package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()
		Convey("When the config values are retrieved", func() {
			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg.APIRouterURL, ShouldEqual, "http://localhost:23200/v1")
				So(cfg.BatchMaxWorkers, ShouldEqual, 100)
				So(cfg.BatchSizeLimit, ShouldEqual, 1000)
				So(cfg.BindAddr, ShouldEqual, "localhost:20001")
				So(cfg.Debug, ShouldBeFalse)
				So(cfg.DownloadServiceURL, ShouldEqual, "http://localhost:23600")
				So(cfg.EnableDatasetPreview, ShouldBeFalse)
				So(cfg.EnableProfiler, ShouldBeFalse)
				So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.MaxDatasetOptions, ShouldEqual, 200)
				So(cfg.PatternLibraryAssetsPath, ShouldEqual, "//cdn.ons.gov.uk/dp-design-system/afa6add")
				So(cfg.SiteDomain, ShouldEqual, "localhost")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
