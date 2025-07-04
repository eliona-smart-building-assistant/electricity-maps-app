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

package app

import (
	"context"
	apiserver "electricity-maps/api/generated"
	apiservices "electricity-maps/api/services"
	appmodel "electricity-maps/app/model"
	"electricity-maps/broker"
	dbhelper "electricity-maps/db/helper"
	"electricity-maps/eliona"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-eliona/dashboard"
	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

var appStatus = 0

const (
	statusOK = iota
	statusError
	statusFatal
)

func changeAppStatus(status int) {
	appStatus = status
	Heartbeat()
}

func Initialize() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// Init the app before the first run.
	app.Init(conn, app.AppName(),
		app.ExecSqlFile("db/init.sql"),
		initAssetCategory(),
		asset.InitAssetTypeFiles("resources/asset-types/*.json"),
		dashboard.InitWidgetTypeFiles("resources/widget-types/*.json"),
	)
}

func initAssetCategory() func(db.Connection) error {
	return func(db.Connection) error {
		_, _, err := client.NewClient().AssetTypesAPI.
			PutAssetTypeCategory(client.AuthenticationContext()).
			AssetTypeCategory(api.AssetTypeCategory{
				Name: "electricity-maps-app-location",
				Translation: *api.NewNullableTranslation(&api.Translation{
					De: api.PtrString("Elektrizitätskarten Standort"),
					En: api.PtrString("Electricity Maps Location"),
				}),
				Properties: []api.AssetTypeCategoryProperty{
					{
						Name: *api.PtrString("name"),
						Translation: *api.NewNullableTranslation(&api.Translation{
							En: api.PtrString("Name"),
							De: api.PtrString("Name"),
						}),
					},
				},
			}).Execute()
		return err
	}
}

var (
	once             sync.Once
	configChangeChan = make(chan struct{})
	previousConfigs  = make(map[int64]appmodel.Configuration)
	configMutex      sync.Mutex
)

func CollectData() {
	config, err := dbhelper.GetConfig(context.Background())
	if errors.Is(err, dbhelper.ErrNotFound) {
		once.Do(func() {
			log.Info("dbhelper", "No configs in DB. Please configure the app in Eliona.")
		})
		return
	}
	if err != nil {
		log.Fatal("dbhelper", "Couldn't read configs from DB: %v", err)
		changeAppStatus(statusFatal)
		return
	}

	if !config.Enable {
		if config.Active {
			dbhelper.SetConfigActiveState(context.Background(), false)
		}
		return
	}

	if !config.Active {
		dbhelper.SetConfigActiveState(context.Background(), true)
		log.Info("dbhelper", "Collecting initialized with Configuration %d:\n"+
			"Enable: %t\n"+
			"Refresh Interval: %d\n"+
			"Request Timeout: %d\n"+
			"Project IDs: %v\n",
			config.Id,
			config.Enable,
			config.RefreshInterval,
			config.RequestTimeout,
			config.ProjectIDs)
	}

	// Check for changes in this specific config
	if isConfigChanged(config) {
		select {
		case configChangeChan <- struct{}{}: // Non-blocking send
			log.Debug("app", "Config changed signal sent")
		default:
			log.Debug("app", "Config change signal not sent, channel full")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	common.RunOnceWithParam(func(config appmodel.Configuration) {
		log.Info("main", "Collecting %d started.", config.Id)
		if err := collectResources(ctx, &config); err != nil {
			changeAppStatus(statusError)
			cancel() // Cancel the context to stop the long-running processes
			return   // Error is handled in the method itself.
		}
		log.Info("main", "Collecting %d finished.", config.Id)
		changeAppStatus(statusOK)

		// Wait for the next interval or a config change
		select {
		case <-time.After(time.Second * time.Duration(config.RefreshInterval)):
			// Continue with the next iteration
			return
		case <-configChangeChan:
			// Config changed, restart the process
			cancel() // Cancel the context to stop the long-running process
			return
		}
	}, config, config.Id)
}

func isConfigChanged(newConfig appmodel.Configuration) bool {
	configMutex.Lock()
	defer configMutex.Unlock()

	oldConfig, exists := previousConfigs[newConfig.Id]
	if !exists {
		// New config added
		previousConfigs[newConfig.Id] = newConfig
		return true
	}

	if !reflect.DeepEqual(newConfig, oldConfig) {
		// Config changed
		previousConfigs[newConfig.Id] = newConfig
		return true
	}

	return false
}

func triggerReload() {
	select {
	case configChangeChan <- struct{}{}:
		log.Debug("app", "Triggered reload via config change signal")
	default:
		log.Debug("app", "Could not trigger reload, channel full")
	}
}

func collectResources(ctx context.Context, config *appmodel.Configuration) error {
	if err := createRootAsset(config); err != nil {
		log.Error("app", "creating root asset for config %v in Eliona: %v", config.Id, err)
		return err
	}

	assets, err := dbhelper.GetAssets(ctx)
	if err != nil {
		log.Error("dbhelper", "getting assets: %v", err)
		return err
	}
	for _, asset := range assets {
		electricityInfo, err := broker.GetZoneData(asset.LocationID, config.ApiKey)
		if err != nil {
			log.Error("broker", "getting electricityInfo data: %v", err)
			return err
		}
		electricityInfoMap := electricityInfoToMap(electricityInfo)
		if err := eliona.UpsertData(asset.AssetID, electricityInfoMap, time.Now(), api.SUBTYPE_INPUT); err != nil {
			log.Error("eliona", "upserting data for asset %v: %v", asset.AssetID, err)
			return err
		}
	}

	return nil
}

func createRootAsset(config *appmodel.Configuration) error {
	if hasRoot, err := dbhelper.RootAssetAlreadyCreated(); err != nil {
		return fmt.Errorf("finding whether config already has root asset: %v", err)
	} else if hasRoot {
		return nil
	}
	var assets []asset.AssetWithParentReferences
	root := eliona.Root{Config: config}
	assets = append(assets, &root)
	if err := eliona.CreateAssets(*config, assets); err != nil {
		return fmt.Errorf("creating assets: %v", err)
	}
	return nil
}

func electricityInfoToMap(info broker.ZoneData) map[string]interface{} {
	attrMap := make(map[string]interface{})
	attrMap["name"] = info.Zone
	attrMap["carbon_intensity"] = info.CarbonIntensity
	attrMap["renewable_percentage"] = info.RenewablePercentage
	attrMap["fossil_free_percentage"] = info.FossilFreePercentage
	return attrMap
}

// ListenForOutputChanges listens to output attribute changes from Eliona. Delete if not needed.
func ListenForOutputChanges() {
	for {
		outputs, err := eliona.ListenForPropertyChanges()
		if err != nil {
			log.Error("eliona", "listening for output changes: %v", err)
			changeAppStatus(statusError)
			return
		}

		for output := range outputs {
			if cr := output.ClientReference.Get(); cr != nil && *cr == eliona.ClientReference {
				continue
			}

			asset, err := dbhelper.GetAssetById(output.AssetId)
			if errors.Is(err, dbhelper.ErrNotFound) {
				handleNewAsset(output)
				triggerReload()
				continue
			} else if err != nil {
				log.Error("dbhelper", "getting asset by assetID %v: %v", output.AssetId, err)
				changeAppStatus(statusError)
				return
			}

			handleExistingAsset(output, asset)
			triggerReload()
		}

		time.Sleep(time.Second * 5)
	}
}

func handleNewAsset(output api.Data) {
	log.Debug("app", "received data update for new asset %v: %+v", output.AssetId, output)

	elionaAsset, err := eliona.GetAsset(output.AssetId)
	if err != nil {
		log.Error("eliona", "getting asset ID %v: %v", output.AssetId, err)
		return
	}

	if elionaAsset.AssetType != "electricity_maps_app_location" {
		log.Debug("eliona", "this asset is not ours")
		return
	}

	locationName, ok := getLocationName(output.Data)
	if !ok {
		return
	}

	config, err := dbhelper.GetConfig(context.Background())
	if err != nil {
		log.Error("dbhelper", "getting config: %v", err)
		changeAppStatus(statusError)
		return
	}

	location, err := broker.Locate(config, locationName)
	if errors.Is(err, broker.ErrNotFound) {
		zones, _ := broker.ListAvailableZones(config.ApiKey)
		msg := "Location not found. Available:"
		for code, zone := range zones {
			msg += fmt.Sprintf(" %s(%s),", zone.ZoneName, code)
		}
		msg = strings.TrimSuffix(msg, ",")
		_ = eliona.UpsertData(elionaAsset.GetId(), map[string]any{"name": msg}, time.Now(), api.SUBTYPE_PROPERTY)
		return
	} else if err != nil {
		log.Warn("app", "trying to locate %s: %v", locationName, err)
		return
	}

	locationNameFormatted := formatLocationName(location)

	if err := eliona.UpsertData(elionaAsset.GetId(), map[string]any{"name": locationNameFormatted}, time.Now(), api.SUBTYPE_PROPERTY); err != nil {
		log.Error("eliona", "updating asset %v location name: %v", elionaAsset.GetId(), err)
		return
	}

	if err := dbhelper.InsertAsset(client.AuthenticationContext(), appmodel.Asset{
		ProjectID:  elionaAsset.ProjectId,
		AssetID:    elionaAsset.GetId(),
		LocationID: location.Code,
	}); err != nil {
		log.Error("dbhelper", "inserting asset: %v", err)
	}
}

func handleExistingAsset(output api.Data, asset appmodel.Asset) {
	log.Debug("app", "received data update for known asset %v: %+v", output.AssetId, output)

	locationName, ok := getLocationName(output.Data)
	if !ok {
		return
	}

	config, err := dbhelper.GetConfig(context.Background())
	if err != nil {
		log.Error("dbhelper", "getting config: %v", err)
		changeAppStatus(statusError)
		return
	}

	location, err := broker.Locate(config, locationName)
	if errors.Is(err, broker.ErrNotFound) {
		zones, _ := broker.ListAvailableZones(config.ApiKey)
		msg := "Location not found. Available:"
		for code, zone := range zones {
			msg += fmt.Sprintf(" %s(%s),", zone.ZoneName, code)
		}
		msg = strings.TrimSuffix(msg, ",")
		_ = eliona.UpsertData(asset.AssetID, map[string]any{"name": msg}, time.Now(), api.SUBTYPE_PROPERTY)
		return
	} else if err != nil {
		log.Warn("app", "trying to locate %s: %v", locationName, err)
		return
	}

	locationNameFormatted := formatLocationName(location)

	if err := eliona.UpsertData(asset.AssetID, map[string]any{"name": locationNameFormatted}, time.Now(), api.SUBTYPE_PROPERTY); err != nil {
		log.Error("eliona", "updating asset %v location name: %v", asset.AssetID, err)
		return
	}

	if err := dbhelper.UpdateAssetLocation(client.AuthenticationContext(), appmodel.Asset{
		ID:         asset.ID,
		LocationID: location.Code,
	}); err != nil {
		log.Error("dbhelper", "updating asset: %v", err)
	}
}

func getLocationName(data map[string]interface{}) (string, bool) {
	outputLocationName, ok := data["name"]
	if !ok {
		log.Warn("eliona", "received known asset type, but don't understand data: %+v", data)
		return "", false
	}

	locationName, ok := outputLocationName.(string)
	if !ok {
		log.Warn("eliona", "received known asset type, but cannot convert location name to string: %+v", data)
		return "", false
	}

	return locationName, true
}

func formatLocationName(location broker.Zone) string {
	return fmt.Sprintf("%s - %s", location.Code, location.ZoneName)
}

// outputData implements passing output data to broker. Remove if not needed.
func outputData(asset appmodel.Asset, data map[string]interface{}) error {
	// Do the output magic here.
	return nil
}

func Heartbeat() {
	roots, err := dbhelper.GetRootAssets()
	if err != nil {
		log.Error("dbhelper", "getting root assets: %v", err)
		return
	}

	for _, root := range roots {
		err := eliona.UpsertData(root.AssetID, map[string]any{"status": appStatus}, time.Now(), api.SUBTYPE_STATUS)
		if err != nil {
			log.Error("eliona", "upserting data as heartbeat: %v", err)
			return
		}
	}
}

// ListenApi starts the API server and listen for requests
func ListenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"),
		frontend.NewEnvironmentHandler(
			utilshttp.NewCORSEnabledHandler(
				apiserver.NewRouter(
					apiserver.NewConfigurationAPIController(apiservices.NewConfigurationAPIService()),
					apiserver.NewVersionAPIController(apiservices.NewVersionAPIService()),
					apiserver.NewCustomizationAPIController(apiservices.NewCustomizationAPIService()),
				))))
	log.Fatal("main", "API server: %v", err)
	changeAppStatus(statusFatal)
}
