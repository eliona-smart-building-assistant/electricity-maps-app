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

package eliona

import (
	"context"
	appmodel "electricity-maps/app/model"
	dbhelper "electricity-maps/db/helper"
)

type Root struct {
	LocationalParentGAI string
	FunctionalParentGAI string

	Config *appmodel.Configuration
}

func (r *Root) GetName() string {
	return "weather_app"
}

func (r *Root) GetDescription() string {
	return "Root asset for Weather App"
}

func (r *Root) GetAssetType() string {
	return "weather_app_root"
}

func (r *Root) GetGAI() string {
	return r.GetAssetType()
}

func (r *Root) GetAssetID(projectID string) (*int32, error) {
	return dbhelper.GetRootAssetId(context.Background(), projectID, r.GetGAI())
}

func (r *Root) SetAssetID(assetID int32, projectID string) error {
	return dbhelper.UpsertRootAsset(assetID, projectID, r.GetGAI())
}

func (r *Root) GetLocationalParentGAI() string {
	return r.LocationalParentGAI
}

func (r *Root) GetFunctionalParentGAI() string {
	return r.FunctionalParentGAI
}
