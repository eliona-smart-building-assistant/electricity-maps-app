# Electricity Maps User Guide

## Introduction
> The Electricity Maps app provides Eliona with real-time electricity grid data from Electricity Maps, including carbon intensity and renewable energy percentages for different geographic zones.

![App schema](https://raw.githubusercontent.com/eliona-smart-building-assistant/electricity-maps-app/refs/heads/develop/schema.png "App schema")

## Overview
This guide provides instructions on configuring, installing, and using the Electricity Maps app to monitor electricity grid composition and carbon intensity data.

## Installation
Install the Electricity Maps app via the Eliona App Store.

## Configuration

### Registering with Electricity Maps
1. Create an account at [Electricity Maps](https://www.electricitymaps.com/)
2. Subscribe to the appropriate API plan (free tier available for basic usage in single region)
3. Generate an API key in your account settings and save it for the Eliona configuration

### Configure the Electricity Maps App
Configurations can be created in Eliona under `Settings > Apps > Electricity Maps` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Use the config endpoint with the PUT method.

Configuration requires the following data:

| Attribute | Description | Required |
|-----------|-------------|----------|
| `apiKey` | Electricity Maps API key obtained in the previous step | Yes |
| `enable` | Flag to enable or disable this configuration | Yes |
| `refreshInterval` | Interval in seconds for data synchronization (minimum 300 recommended) | Yes |
| `requestTimeout` | API query timeout in seconds | No (default: 120) |
| `projectIDs` | List of Eliona project IDs for data collection | Yes |

Example configuration JSON:
```json
{
  "apiKey": "your-api-key-here",
  "enable": true,
  "refreshInterval": 900,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```
## Asset Creation
Once configured, the app creates an `Electricity Zone` asset type. You can create multiple assets of this type, each representing a geographic zone to monitor.

## Configuring Electricity Zone Locations
1. Create a new asset of type `Electricity Zone`
2. Click the edit button on the asset
3. In the "more info" section, set the zone identifier (e.g., "CH" or "Switzerland" for Switzerland, "DE" or "Germany" for Germany) 
4. Save the asset configuration
5. Refresh the page to verify the app has correctly identified the zone

The asset will then be populated with electricity grid data:

| Attribute | Description | Unit |
|-----------|-------------|------|
| carbon_intensity | Carbon intensity of electricity consumption | gCO₂eq/kWh |
| renewable_percentage | Percentage of renewable energy in electricity consumption | % |
| fossil_free_percentage | Percentage of fossil-free energy in electricity consumption | % |

## App Status Monitoring
The app creates a root asset called "Electricity Maps Root" which provides information about the app's status:

- Asset status: Active/Inactive indicates if the app is running
- Status attribute: Shows the current operational status. If the app status is not "OK", it signifies that the app might not be functioning properly. If the error state persists, let us know by submitting a bug report.

## Use Cases
The Electricity Maps app enables:

- Real-time monitoring of grid carbon intensity
- Tracking renewable energy penetration
- Energy procurement optimization
- ESG reporting and sustainability tracking
- Demand response strategies based on grid composition
- Carbon-aware load shifting
