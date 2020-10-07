package antenna

import (
	"testing"

	contractsv1 "github.com/JorritSalverda/jarvis-contracts-golang/contracts/v1"
	apiv1 "github.com/JorritSalverda/jarvis-uponor-smatrix-exporter/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestGetMeasurement(t *testing.T) {
	t.Run("ReturnsMeasurement", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client, err := NewClient("/dev/ttyUSB0")
		assert.Nil(t, err)

		config := apiv1.Config{
			Location: "My address",
		}

		// act
		measurement, err := client.GetMeasurement(config)

		assert.Nil(t, err)
		assert.Equal(t, "My address", measurement.Location)
	})

	t.Run("ReturnsMeasurementWithSample", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client, err := NewClient("/dev/ttyUSB0")
		assert.Nil(t, err)

		config := apiv1.Config{
			Location: "My address",
			SampleConfigs: []apiv1.ConfigSample{
				{
					EntityType:      "ENTITY_TYPE_ZONE",
					EntityName:      "Uponor Smatrix",
					SampleType:      "SAMPLE_TYPE_TEMPERATURE",
					SampleName:      "Living room",
					MetricType:      "METRIC_TYPE_GAUGE",
					ValueMultiplier: 1,
					ThermostatID:    "abcd",
				},
			},
		}

		// act
		measurement, err := client.GetMeasurement(config)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(measurement.Samples))
		assert.Equal(t, "Uponor Smatrix", measurement.Samples[0].EntityName)
		assert.Equal(t, "Living room", measurement.Samples[0].SampleName)
		assert.Equal(t, contractsv1.MetricType_METRIC_TYPE_GAUGE, measurement.Samples[0].MetricType)
	})
}
