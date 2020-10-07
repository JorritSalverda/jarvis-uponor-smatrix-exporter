package config

import (
	"context"
	"testing"

	contractsv1 "github.com/JorritSalverda/jarvis-contracts-golang/contracts/v1"
	"github.com/stretchr/testify/assert"
)

func TestReadConfigFromFile(t *testing.T) {

	t.Run("ReturnsConfig", func(t *testing.T) {

		ctx := context.Background()
		client, _ := NewClient(ctx)

		// act
		config, err := client.ReadConfigFromFile("./test-config.yaml")

		assert.Nil(t, err)
		assert.Equal(t, "My Home", config.Location)
		assert.Equal(t, 2, len(config.SampleConfigs))
		assert.Equal(t, contractsv1.EntityType_ENTITY_TYPE_ZONE, config.SampleConfigs[0].EntityType)
		assert.Equal(t, "Uponor Smatrix T-169", config.SampleConfigs[0].EntityName)
		assert.Equal(t, contractsv1.SampleType_SAMPLE_TYPE_TEMPERATURE, config.SampleConfigs[0].SampleType)
		assert.Equal(t, "Bathroom", config.SampleConfigs[0].SampleName)
		assert.Equal(t, contractsv1.MetricType_METRIC_TYPE_GAUGE, config.SampleConfigs[0].MetricType)
	})
}
