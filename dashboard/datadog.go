package dashboard

import (
	"PRism/config"
	"context"
	"encoding/json"
	"fmt"
	"log"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

func CreateDatadogDashboard(suggestion config.DashboardSuggestion, cfg config.Config) error {
	log.Printf("Creating Datadog dashboard: %s", suggestion.Name)

	// Initialize Datadog client with the v1 API client
	configuration := datadog.NewConfiguration()
	configuration.Host = "api.ap1.datadoghq.com"
	configuration.AddDefaultHeader("DD-API-KEY", cfg.DatadogAPIKey)
	configuration.AddDefaultHeader("DD-APPLICATION-KEY", cfg.DatadogAppKey)
	apiClient := datadog.NewAPIClient(configuration)

	// Parse the queries, panels, and alerts
	var queries []map[string]interface{}
	var panels []map[string]interface{}
	var alerts []map[string]interface{}

	if err := json.Unmarshal([]byte(suggestion.Queries), &queries); err != nil {
		return fmt.Errorf("error parsing queries JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(suggestion.Panels), &panels); err != nil {
		return fmt.Errorf("error parsing panels JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(suggestion.Alerts), &alerts); err != nil {
		return fmt.Errorf("error parsing alerts JSON: %v", err)
	}

	// Create widgets from panels
	widgets := []datadog.Widget{}
	for _, panel := range panels {
		title := panel["title"].(string)
		gridPos := panel["gridPos"].(map[string]interface{})
		targets := panel["targets"].([]interface{})

		// Create widget requests
		requests := []datadog.TimeseriesWidgetRequest{}
		for _, target := range targets {
			targetID := target.(map[string]interface{})["refId"].(string)
			for _, query := range queries {
				if query["refId"].(string) == targetID {
					queryStr := query["expr"].(string)

					// Create a timeserieswidgetrequest
					request := datadog.TimeseriesWidgetRequest{
						Q:           &queryStr,
						DisplayType: datadog.WIDGETDISPLAYTYPE_LINE.Ptr(),
						Style: &datadog.WidgetRequestStyle{
							Palette:   (*string)(datadog.WIDGETPALETTE_BLACK_ON_LIGHT_GREEN.Ptr()),
							LineType:  datadog.WIDGETLINETYPE_SOLID.Ptr(),
							LineWidth: datadog.WIDGETLINEWIDTH_NORMAL.Ptr(),
						},
					}
					requests = append(requests, request)
				}
			}
		}

		// Extract layout parameters
		x := int64(gridPos["x"].(float64))
		y := int64(gridPos["y"].(float64))
		w := int64(gridPos["w"].(float64))
		h := int64(gridPos["h"].(float64))

		// Create widget definition
		legendSize := "small"
		timeseriesDef := datadog.NewTimeseriesWidgetDefinitionWithDefaults()
		timeseriesDef.SetRequests(requests)
		timeseriesDef.SetTitle(title)
		timeseriesDef.SetLegendSize(legendSize)

		// Create widget with definition and layout
		widget := datadog.Widget{
			Definition: datadog.WidgetDefinition{
				TimeseriesWidgetDefinition: timeseriesDef,
			},
			Layout: &datadog.WidgetLayout{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			},
		}
		widgets = append(widgets, widget)
	}

	// Create dashboard template variables
	defaultVal := "*"
	prefix := "env"
	name := "env"
	templateVar := datadog.DashboardTemplateVariable{
		Name:    name,
		Prefix:  *datadog.NewNullableString(&prefix),
		Default: *datadog.NewNullableString(&defaultVal),
	}

	// Create dashboard request
	dashTitle := suggestion.Name
	dashDesc := "Created by PRism"
	layoutType := datadog.DASHBOARDLAYOUTTYPE_ORDERED
	dashboardRequest := datadog.Dashboard{
		Title:             dashTitle,
		Description:       *datadog.NewNullableString(&dashDesc),
		LayoutType:        layoutType,
		Widgets:           widgets,
		TemplateVariables: []datadog.DashboardTemplateVariable{templateVar},
		NotifyList:        []string{},
	}

	// Create the dashboard
	ctx := context.Background()
	dashboard, _, err := apiClient.DashboardsApi.CreateDashboard(ctx).Body(dashboardRequest).Execute()
	if err != nil {
		return fmt.Errorf("failed to create Datadog dashboard: %w", err)
	}

	log.Printf("Successfully created Datadog dashboard with ID: %s", dashboard.GetId())
	return nil
}
