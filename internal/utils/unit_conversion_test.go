package utils

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UnitConversionTestSuite struct {
	suite.Suite
}

func TestUnitConversionTestSuite(t *testing.T) {
	suite.Run(t, new(UnitConversionTestSuite))
}

func (suite *UnitConversionTestSuite) TestUnitConversionFunctionsReturnCorrectValues() {
	type testCase struct {
		function  func(float32) float32
		withValue float32
		wantValue float32
	}

	// Arrange
	testCases := []testCase{
		{BarToPSI, 1, 14.50377},
		{BarToInHg, 1, 29.52998},
		{BarToKPA, 1, 100},
		{CelsiusToFahrenheit, 1, 33.8},
		{MetersToFeet, 1, 3.28084},
		{MetersToInches, 1, 39.3701},
		{MetersToMillimeters, 1, 1000},
		{MetersPerSecondToKilometersPerHour, 1, 3.6},
		{MetersPerSecondToMilesPerHour, 1, 2.2369363},
		{RadiansPerSecondToRevolutionsPerMinute, 1, 9.549296},
	}

	for _, tc := range testCases {
		fnNameSegments := strings.Split(runtime.FuncForPC(reflect.ValueOf(tc.function).Pointer()).Name(), ".")
		fnName := fnNameSegments[len(fnNameSegments)-1]

		suite.Run(fnName, func() {
			// Act
			gotValue := tc.function(tc.withValue)

			// Assert
			suite.Equal(tc.wantValue, gotValue)
		})
	}
}
