package providers

import (
	"math"
	"testing"
	"testing/quick"
)

func TestValidateTemperaturePropertyBounds(t *testing.T) {
	property := func(raw int16) bool {
		temp := float64(raw) / 100
		err := ValidateTemperature(temp)
		if math.IsNaN(temp) {
			return err != nil
		}
		if temp < 0 || temp > 2 {
			return err != nil
		}
		return err == nil
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func TestValidateMaxTokensPropertyPositive(t *testing.T) {
	property := func(tokens int16) bool {
		err := ValidateMaxTokens(int(tokens))
		if tokens <= 0 {
			return err != nil
		}
		return err == nil
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}
