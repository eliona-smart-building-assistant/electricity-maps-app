# Electricity Maps User Guide

### Introduction

> The Electricity Maps app provides integration and synchronization between Eliona and Electricity Maps services.

## Overview

This guide provides instructions on configuring, installing, and using the Electricity Maps app to manage resources and synchronize data between Eliona and Electricity Maps services.

## Installation

Install the Electricity Maps app via the Eliona App Store.

## Configuration

The Electricity Maps app requires configuration through Eliona’s settings interface. Below are the general steps and details needed to configure the app.

### Registering the app in Electricity Maps Service

Create credentials in Electricity Maps Service to connect the Electricity Maps services from Eliona. All required credentials are listed below in the [configuration section](#configure-the-electricity-maps-app).

<mark>TODO: Describe the steps where you can get or create the necessary credentials.</mark>

### Configure the Electricity Maps app

Configurations can be created in Eliona under `Settings > Apps > Electricity Maps` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the appropriate endpoint with the POST method. Each configuration requires the following data:

| Attribute         | Description                                                                     |
|-------------------|---------------------------------------------------------------------------------|
| `baseURL`         | URL of the Electricity Maps services.                                                   |
| `clientSecrets`   | Client secrets obtained from the Electricity Maps service.                              |
| `assetFilter`     | Filtering asset during [Continuous Asset Creation](#continuous-asset-creation). |
| `enable`          | Flag to enable or disable this configuration.                                   |
| `refreshInterval` | Interval in seconds for data synchronization.                                   |
| `requestTimeout`  | API query timeout in seconds.                                                   |
| `projectIDs`      | List of Eliona project IDs for data collection.                                 |

Example configuration JSON:

```json
{
  "baseURL": "http://service/v1",
  "clientSecrets": "random-cl13nt-s3cr3t",
  "filter": "",
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

## Continuous Asset Creation

Once configured, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and users are notified via Eliona’s notification system.

<mark>TODO: Describe what resources are created, the hierarchy and the data points.</mark>

### Asset filtering

In case it's not desired to import all assets from Electricity Maps to Eliona, it's possible to write an asset filter that would include only matching assets. This app is able to filter the assets by: <mark>TODO</mark>. See [Asset Filter documentation](https://doc.eliona.io/collection/eliona-english/manuals/settings/apps/asset-filter) for instructions on writing asset filters.

## Additional Features

<mark>TODO: Describe all other features of the app.</mark>

### Dashboard templates

The app offers a predefined dashboard that clearly displays the most important information. You can create such a dashboard under `Dashboards > Copy Dashboard > From App > Electricity Maps`.

### <mark>TODO: Other features</mark>

## App status monitoring

Along with asset creation, an asset called "Electricity Maps root" is also created. It's purpose is to inform users of the app status -- It signalizes whether the app is running (Asset status -> Active/Inactive) and it's status - the Status attribute. If the app status is not "OK", it signifies that the app might not be functioning properly. If the error state persists, let us know by submitting a bug report.
