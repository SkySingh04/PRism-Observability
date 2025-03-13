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
		title, ok := panel["title"].(string)
		if !ok {
			log.Printf("Warning: panel title is not a string, skipping panel")
			continue
		}

		gridPos, ok := panel["gridPos"].(map[string]interface{})
		if !ok {
			log.Printf("Warning: gridPos is not a map, skipping panel %s", title)
			continue
		}

		targets, ok := panel["targets"].([]interface{})
		if !ok {
			log.Printf("Warning: targets is not an array, skipping panel %s", title)
			continue
		}

		// Create widget requests
		requests := []datadog.TimeseriesWidgetRequest{}
		for _, target := range targets {
			targetMap, ok := target.(map[string]interface{})
			if !ok {
				log.Printf("Warning: target is not a map in panel %s, skipping target", title)
				continue
			}

			refId, ok := targetMap["refId"].(string)
			if !ok {
				log.Printf("Warning: refId is not a string in panel %s, skipping target", title)
				continue
			}

			for _, query := range queries {
				queryRefId, ok := query["refId"].(string)
				if !ok {
					continue
				}

				if queryRefId == refId {
					queryExpr, ok := query["expr"].(string)
					if !ok {
						log.Printf("Warning: expr is not a string in query %s, skipping query", queryRefId)
						continue
					}

					// Create a timeserieswidgetrequest
					request := datadog.TimeseriesWidgetRequest{
						Q:           &queryExpr,
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

		// Extract layout parameters with type checking
		x, ok := getInt64FromFloat(gridPos, "x")
		if !ok {
			log.Printf("Warning: x is not a number in panel %s, using default 0", title)
			x = 0
		}

		y, ok := getInt64FromFloat(gridPos, "y")
		if !ok {
			log.Printf("Warning: y is not a number in panel %s, using default 0", title)
			y = 0
		}

		w, ok := getInt64FromFloat(gridPos, "w")
		if !ok {
			log.Printf("Warning: w is not a number in panel %s, using default 12", title)
			w = 12
		}

		h, ok := getInt64FromFloat(gridPos, "h")
		if !ok {
			log.Printf("Warning: h is not a number in panel %s, using default 8", title)
			h = 8
		}

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
	dashboard, _, err := apiClient.DashboardsApi.CreateDashboard(ctx, dashboardRequest)
	fmt.Println(dashboard)
	if err != nil {
		log.Printf("Failed to create Datadog dashboard: %v", err)
		return fmt.Errorf("failed to create Datadog dashboard: %w", err)
	}

	log.Printf("Successfully created Datadog dashboard with ID: %s", dashboard.GetId())
	return nil
}

// Helper function to safely convert interface{} to int64
func getInt64FromFloat(m map[string]interface{}, key string) (int64, bool) {
	val, exists := m[key]
	if !exists {
		return 0, false
	}

	// Try as float64 (common for JSON numbers)
	if floatVal, ok := val.(float64); ok {
		return int64(floatVal), true
	}

	// Try as int
	if intVal, ok := val.(int); ok {
		return int64(intVal), true
	}

	// Try as int64
	if int64Val, ok := val.(int64); ok {
		return int64Val, true
	}

	// Try as string
	if strVal, ok := val.(string); ok {
		var result float64
		if _, err := fmt.Sscanf(strVal, "%f", &result); err == nil {
			return int64(result), true
		}
	}

	return 0, false
}
