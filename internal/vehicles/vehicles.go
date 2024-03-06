package vehicles

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Vehicle struct {
	ID           int
	Model        string
	Manufacturer string
	Year         int
	Category     string
	CarType      string
	Drivetrain   string
	Aspiration   string
	OpenCockpit  bool
}

type Inventory struct {
	db map[string]Vehicle
}

//go:embed inventory.json
var baseInventoryJSON []byte

func NewInventory(file string) (*Inventory, error) {
	inventory := Inventory{}

	jsonData := baseInventoryJSON
	if file != "" {
		var err error
		jsonData, err = os.ReadFile(file)
		if err != nil {
			fmt.Printf("failed to read file: %s\n", err)
			return &Inventory{}, err
		}
	}

	err := json.Unmarshal([]byte(jsonData), &inventory.db)
	if err != nil {
		fmt.Printf("failed to unmarshal json: %s\n", err)
		return &Inventory{}, err
	}

	return &inventory, nil
}

func (i *Inventory) GetVehicleByID(id int) (Vehicle, error) {
	vehicle, ok := i.db[strconv.Itoa(id)]
	if !ok {
		return Vehicle{}, fmt.Errorf("vehicle with id %d not found", id)
	}

	return vehicle, nil
}

func (v *Vehicle) ExpandedAspiration() string {
	switch v.Aspiration {
	case "NA":
		return "Naturally Aspirated"
	case "TC":
		return "Turbocharged"
	case "SC":
		return "Supercharged"
	case "TC+SC":
		return "Compound Charged"
	default:
		return v.Aspiration
	}
}
