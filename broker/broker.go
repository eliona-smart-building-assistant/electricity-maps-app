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
		if distance > 0 && distance < bestDistance {
			bestDistance = distance
			bestMatch = zone
			if bestDistance == 1 { // Perfect match
				return bestMatch, nil
			}
		}

		// Check ZoneName
		distance = fuzzy.RankMatchNormalizedFold(searchTerm, zone.ZoneName)
		if distance > 0 && distance < bestDistance {
			bestDistance = distance
			bestMatch = zone
			if bestDistance == 1 { // Perfect match
				return bestMatch, nil
			}
		}
	}

	if bestDistance < 5 { // Return if we have a reasonably good match
		return bestMatch, nil
	}

	return Zone{}, ErrNotFound
}
