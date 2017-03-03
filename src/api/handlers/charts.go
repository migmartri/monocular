package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/helm/monocular/src/api/chartpackagesort"
	"github.com/helm/monocular/src/api/data"
	"github.com/helm/monocular/src/api/data/helpers"
	"github.com/helm/monocular/src/api/swagger/models"
	chartsapi "github.com/helm/monocular/src/api/swagger/restapi/operations/charts"
)

const (
	// ChartResourceName is the resource type string for a chart
	ChartResourceName = "chart"
	// ChartVersionResourceName is the resource type string for a chart version
	ChartVersionResourceName = "chartVersion"
)

// GetChart is the handler for the /charts/{repo}/{name} endpoint
func GetChart(params chartsapi.GetChartParams, c data.Charts) middleware.Responder {
	chartPackage, err := c.ChartFromRepo(params.Repo, params.ChartName)
	if err != nil {
		log.Printf("data.chartsapi.ChartFromRepo(%s, %s) error (%s)", params.Repo, params.ChartName, err)
		return notFound(ChartResourceName)
	}
	chartResource := helpers.MakeChartResource(chartPackage)
	return chartHTTPBody(chartResource)
}

// GetChartVersion is the handler for the /charts/{repo}/{name}/versions/{version} endpoint
func GetChartVersion(params chartsapi.GetChartVersionParams, c data.Charts) middleware.Responder {
	chartPackage, err := c.ChartVersionFromRepo(params.Repo, params.ChartName, params.Version)
	if err != nil {
		log.Printf("data.chartsapi.ChartVersionFromRepo(%s, %s, %s) error (%s)", params.Repo, params.ChartName, params.Version, err)
		return notFound(ChartVersionResourceName)
	}
	chartVersionResource := helpers.MakeChartVersionResource(chartPackage)
	return chartHTTPBody(chartVersionResource)
}

// GetChartVersions is the handler for the /charts/{repo}/{name}/versions endpoint
func GetChartVersions(params chartsapi.GetChartVersionsParams, c data.Charts) middleware.Responder {
	chartPackages, err := c.ChartVersionsFromRepo(params.Repo, params.ChartName)
	if err != nil {
		log.Printf("data.chartsapi.ChartVersionsFromRepo(%s, %s) error (%s)", params.Repo, params.ChartName, err)
		return notFound(ChartVersionResourceName)
	}

	// Sort by semver reverse order
	sort.Sort(sort.Reverse(chartpackagesort.BySemver(chartPackages)))

	chartVersionResources := helpers.MakeChartVersionResources(chartPackages)
	return chartsHTTPBody(chartVersionResources)
}

// GetAllCharts is the handler for the /charts endpoint
func GetAllCharts(params chartsapi.GetAllChartsParams, c data.Charts) middleware.Responder {
	charts, err := c.All()
	if err != nil {
		log.Printf("data.Charts All() error (%s)", err)
		return notFound(ChartResourceName + "s")
	}

	// For now we only sort by name
	sort.Sort(chartpackagesort.ByName(charts))
	resources := helpers.MakeChartResources(charts)
	return chartsHTTPBody(resources)
}

// GetChartsInRepo is the handler for the /charts/{repo} endpoint
func GetChartsInRepo(params chartsapi.GetChartsInRepoParams, c data.Charts) middleware.Responder {
	charts, err := c.AllFromRepo(params.Repo)
	if err != nil {
		log.Printf("data.Charts AllFromRepo(%s) error (%s)", params.Repo, err)
		return notFound(ChartResourceName + "s")
	}
	// For now we only sort by name
	sort.Sort(chartpackagesort.ByName(charts))
	chartsResource := helpers.MakeChartResources(charts)
	return chartsHTTPBody(chartsResource)
}

// SearchCharts is the handler for the /charts/search endpoint
func SearchCharts(params chartsapi.SearchChartsParams, c data.Charts) middleware.Responder {
	charts, err := c.Search(params)
	if err != nil {
		message := fmt.Sprintf("data.Charts Search() error (%s)", err)
		log.Printf(message)
		return chartsapi.NewSearchChartsDefault(http.StatusBadRequest).WithPayload(
			&models.Error{Code: helpers.Int64ToPtr(http.StatusNotFound), Message: &message},
		)
	}
	resources := helpers.MakeChartResources(charts)
	return chartsHTTPBody(resources)
}

// chartHTTPBody is a convenience that returns a swagger-friendly HTTP 200 response with chart body data
func chartHTTPBody(chart *models.Resource) middleware.Responder {
	resourceData := models.ResourceData{
		Data: chart,
	}
	return chartsapi.NewGetChartOK().WithPayload(&resourceData)
}

// chartsHTTPBody is a convenience that returns a swagger-friendly HTTP 200 response with charts body data
func chartsHTTPBody(charts []*models.Resource) middleware.Responder {
	resourceArrayData := models.ResourceArrayData{
		Data: charts,
	}
	return chartsapi.NewGetAllChartsOK().WithPayload(&resourceArrayData)
}
