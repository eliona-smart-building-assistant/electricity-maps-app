//  This file is part of the Eliona project.
//  Copyright © 2025 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package broker

import (
	appmodel "electricity-maps/app/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var ErrNotFound = errors.New("not found")

// TestAuthentication tests if the provided API key is valid
func TestAuthentication(config appmodel.Configuration) error {
	_, err := getZones(config.ApiKey)
	return err
}

// Zone represents a geographical zone with electricity data access
type Zone struct {
	Code     string
	ZoneName string   `json:"zoneName"`
	Access   []string `json:"access"`
}

// zoneResponse represents the response from the zones endpoint
type zoneResponse map[string]Zone

func getZones(apiKey string) (zoneResponse, error) {
	url := "https://api.electricitymap.org/v3/zones"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("auth-token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		// Check for error response
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("API error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("unsuccessful response: %s: %v", resp.Status, string(body))
	}

	var zoneResponse zoneResponse
	err = json.Unmarshal(body, &zoneResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return zoneResponse, nil
}

// Locate finds a zone by its ID or name with fuzzy matching
func Locate(config appmodel.Configuration, name string) (Zone, error) {
	zones, err := getZones(config.ApiKey)
	if err != nil {
		return Zone{}, fmt.Errorf("getting zones: %w", err)
	}

	// Try to find the best match
	var bestMatch Zone
	bestDistance := 1000
	searchTerm := strings.ToLower(name)

	for id, zone := range zones {
		zone.Code = id
		// Check ID
		distance := fuzzy.RankMatchNormalizedFold(searchTerm, id)
		if distance >= 0 && distance < bestDistance {
			bestDistance = distance
			bestMatch = zone
			if bestDistance == 0 { // Perfect match
				return bestMatch, nil
			}
		}

		// Check ZoneName
		distance = fuzzy.RankMatchNormalizedFold(searchTerm, zone.ZoneName)
		if distance >= 0 && distance < bestDistance {
			bestDistance = distance
			bestMatch = zone
			if bestDistance == 0 { // Perfect match
				return bestMatch, nil
			}
		}
	}

	if bestDistance < 5 { // Return if we have a reasonably good match
		return bestMatch, nil
	}

	return Zone{}, ErrNotFound
}

// ListAvailableZones returns all available zones from the Electricity Maps API
func ListAvailableZones(apiKey string) (map[string]Zone, error) {
	zones, err := getZones(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	return zones, nil
}

// PowerBreakdown represents the detailed power breakdown data
type PowerBreakdown struct {
	Nuclear          *float64 `json:"nuclear"`
	Geothermal       *float64 `json:"geothermal"`
	Biomass          *float64 `json:"biomass"`
	Coal             *float64 `json:"coal"`
	Wind             *float64 `json:"wind"`
	Solar            *float64 `json:"solar"`
	Hydro            *float64 `json:"hydro"`
	Gas              *float64 `json:"gas"`
	Oil              *float64 `json:"oil"`
	Unknown          *float64 `json:"unknown"`
	HydroDischarge   *float64 `json:"hydro discharge"`
	BatteryDischarge *float64 `json:"battery discharge"`
}

// ZoneData represents the combined electricity data for a zone
type ZoneData struct {
	Zone                      string             `json:"zone"`
	CarbonIntensity           float64            `json:"carbonIntensity"`
	Datetime                  time.Time          `json:"datetime"`
	UpdatedAt                 time.Time          `json:"updatedAt"`
	CreatedAt                 time.Time          `json:"createdAt"`
	EmissionFactorType        string             `json:"emissionFactorType"`
	IsEstimated               bool               `json:"isEstimated"`
	EstimationMethod          string             `json:"estimationMethod"`
	PowerConsumptionBreakdown PowerBreakdown     `json:"powerConsumptionBreakdown"`
	PowerProductionBreakdown  PowerBreakdown     `json:"powerProductionBreakdown"`
	PowerImportBreakdown      map[string]float64 `json:"powerImportBreakdown"`
	PowerExportBreakdown      map[string]float64 `json:"powerExportBreakdown"`
	FossilFreePercentage      float64            `json:"fossilFreePercentage"`
	RenewablePercentage       float64            `json:"renewablePercentage"`
	PowerConsumptionTotal     float64            `json:"powerConsumptionTotal"`
	PowerProductionTotal      float64            `json:"powerProductionTotal"`
	PowerImportTotal          float64            `json:"powerImportTotal"`
	PowerExportTotal          float64            `json:"powerExportTotal"`
}

// GetZoneData retrieves comprehensive electricity data for a specific zone
func GetZoneData(zone string, apiKey string) (ZoneData, error) {
	// First get carbon intensity data
	carbonURL := fmt.Sprintf("https://api.electricitymap.org/v3/carbon-intensity/latest?zone=%s", zone)
	carbonData, err := fetchData[carbonIntensityResponse](carbonURL, apiKey)
	if err != nil {
		return ZoneData{}, fmt.Errorf("failed to get carbon intensity: %w", err)
	}

	// Then get power breakdown data
	powerURL := fmt.Sprintf("https://api.electricitymap.org/v3/power-breakdown/latest?zone=%s", zone)
	powerData, err := fetchData[powerBreakdownResponse](powerURL, apiKey)
	if err != nil {
		return ZoneData{}, fmt.Errorf("failed to get power breakdown: %w", err)
	}

	// Merge the data into a single response
	zoneData := ZoneData{
		Zone:                      carbonData.Zone,
		CarbonIntensity:           carbonData.CarbonIntensity,
		Datetime:                  carbonData.Datetime,
		UpdatedAt:                 carbonData.UpdatedAt,
		CreatedAt:                 carbonData.CreatedAt,
		EmissionFactorType:        carbonData.EmissionFactorType,
		IsEstimated:               carbonData.IsEstimated,
		EstimationMethod:          carbonData.EstimationMethod,
		PowerConsumptionBreakdown: powerData.PowerConsumptionBreakdown,
		PowerProductionBreakdown:  powerData.PowerProductionBreakdown,
		PowerImportBreakdown:      powerData.PowerImportBreakdown,
		PowerExportBreakdown:      powerData.PowerExportBreakdown,
		FossilFreePercentage:      powerData.FossilFreePercentage,
		RenewablePercentage:       powerData.RenewablePercentage,
		PowerConsumptionTotal:     powerData.PowerConsumptionTotal,
		PowerProductionTotal:      powerData.PowerProductionTotal,
		PowerImportTotal:          powerData.PowerImportTotal,
		PowerExportTotal:          powerData.PowerExportTotal,
	}

	return zoneData, nil
}

type carbonIntensityResponse struct {
	Zone               string    `json:"zone"`
	CarbonIntensity    float64   `json:"carbonIntensity"`
	Datetime           time.Time `json:"datetime"`
	UpdatedAt          time.Time `json:"updatedAt"`
	CreatedAt          time.Time `json:"createdAt"`
	EmissionFactorType string    `json:"emissionFactorType"`
	IsEstimated        bool      `json:"isEstimated"`
	EstimationMethod   string    `json:"estimationMethod"`
}

type powerBreakdownResponse struct {
	Zone                      string             `json:"zone"`
	Datetime                  time.Time          `json:"datetime"`
	UpdatedAt                 time.Time          `json:"updatedAt"`
	CreatedAt                 time.Time          `json:"createdAt"`
	PowerConsumptionBreakdown PowerBreakdown     `json:"powerConsumptionBreakdown"`
	PowerProductionBreakdown  PowerBreakdown     `json:"powerProductionBreakdown"`
	PowerImportBreakdown      map[string]float64 `json:"powerImportBreakdown"`
	PowerExportBreakdown      map[string]float64 `json:"powerExportBreakdown"`
	FossilFreePercentage      float64            `json:"fossilFreePercentage"`
	RenewablePercentage       float64            `json:"renewablePercentage"`
	PowerConsumptionTotal     float64            `json:"powerConsumptionTotal"`
	PowerProductionTotal      float64            `json:"powerProductionTotal"`
	PowerImportTotal          float64            `json:"powerImportTotal"`
	PowerExportTotal          float64            `json:"powerExportTotal"`
	IsEstimated               bool               `json:"isEstimated"`
	EstimationMethod          string             `json:"estimationMethod"`
}

func fetchData[T any](url string, apiKey string) (T, error) {
	var empty T

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return empty, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("auth-token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return empty, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			return empty, fmt.Errorf("API error: %s", errorResp.Error)
		}
		return empty, fmt.Errorf("unsuccessful response: %s: %s", resp.Status, string(body))
	}

	var result T
	err = json.Unmarshal(body, &result)
	if err != nil {
		return empty, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}
