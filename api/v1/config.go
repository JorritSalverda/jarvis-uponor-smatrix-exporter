package api

import (
	contractsv1 "github.com/JorritSalverda/jarvis-contracts-golang/contracts/v1"
)

type Config struct {
	Location      string         `yaml:"location"`
	SampleConfigs []ConfigSample `yaml:"sampleConfigs"`
}

type ConfigSample struct {
	// default jarvis config for sample
	EntityType contractsv1.EntityType `yaml:"entityType"`
	EntityName string                 `yaml:"entityName"`
	SampleType contractsv1.SampleType `yaml:"sampleType"`
	SampleName string                 `yaml:"sampleName"`
	MetricType contractsv1.MetricType `yaml:"metricType"`

	// alpha innotec specific config for sample
	ValueMultiplier float64 `yaml:"valueMultiplier"`
	ThermostatID    string  `yaml:"thermostatID"`
}

func (c *Config) SetDefaults() {
	for _, sc := range c.SampleConfigs {
		sc.SetDefaults()
	}
}

func (sc *ConfigSample) SetDefaults() {
	if sc.ValueMultiplier == 0 {
		sc.ValueMultiplier = 1
	}
}
