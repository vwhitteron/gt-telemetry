# GT Telemetry #

[![Build Status](https://github.com/vwhitteron/gt-telemetry/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/vwhitteron/gt-telemetry/actions?query=branch%3Amain)
[![codecov](https://codecov.io/gh/vwhitteron/gt-telemetry/branch/main/graph/badge.svg)](https://codecov.io/gh/vwhitteron/gt-telemetry)
[![Go Report Card](https://goreportcard.com/badge/github.com/vwhitteron/gt-telemetry)](https://goreportcard.com/report/github.com/vwhitteron/gt-telemetry)

GT Telemetry is a module for reading Gran Turismo race telemetry streams in Go.

## Features

* Support for all fields contained within the telemetry data packet.
* Access data in both metric and imperial units.
* An additional field for the differential gear ratio is computed based on the rolling wheel diameter of the driven wheels.
* A vehicle inventory database for providing the follwing information on a given vehicle ID:
  * Manufacturer
  * Model
  * Year
  * Drivetrain
  * Aspiration
  * Type (racing or street)
  * Racing category
  * Open cockpit exposure


## Known issues ##

* The [vehicle inventory database](https://github.com/vwhitteron/gt-telemetry/internal/vehicles/inventory.json) is missing information on some vehicles. Feel free to raise a pull request to add any missing or incorrect information.
* The differential ratio may be incorrect for vehicles that are not defined in the inventory database. By default the ratio is based on the rolling diameter of the rear wheels so will result in an incorrect ratio for front wheel drive cars with staggered wheel diameters.


## Installation ##

To start using gt-telemetry, install Go 1.21 or above. From your project, run the following command to retrieve the module:

```bash
go get github.com/vwhitteron/gt-telemetry
```

## Usage ##

```go
import telemetry_client "github.com/vwhitteron/gt-telemetry"
```

Construct a new GT client and start reading the telemetry stream. All configuration fields are optional with the default values show in the example.

```go
config := telemetry_client.Config{
    IPAddr: "255.255.255.255",
    LogLevel: "info",
    StatsEnabled: false,
    VehicleDB: "./internal/vehicles/inventory.json",
}
gt, _ := telemetry_client.NewGTClient(config)
go gt.Run()
```

_If the PlayStation is on the same network segment then you will probably find that the default broadcast address `255.255.255.255` will be sufficient to start reading data. If it does not work then enter the IP address of the PlayStation device instead._

Read some data from the stream:

```go
    fmt.Printf("Sequence ID:  %6d    %3.0f kph  %5.0f rpm\n",
        gt.Telemetry.SequenceID(),
        gt.Telemetry.GroundSpeedKPH(),
        gt.Telemetry.EngineRPM(),
    )
```

## Examples ##

The [examples](./examples) directory contains an example for accessing most data made available by the library. The telemetry data can be viewed by running:

```bash
make run/live
```

## Acknowledgements ##
Special thanks to [Nenkai](https://github.com/Nenkai) for the excellent work documenting the Gran Turismo telemetry protocol.