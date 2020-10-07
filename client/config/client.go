package config

import (
	"context"
	"io/ioutil"

	apiv1 "github.com/JorritSalverda/jarvis-uponor-smatrix-exporter/api/v1"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type Client interface {
	ReadConfigFromFile(path string) (config apiv1.Config, err error)
}

func NewClient(ctx context.Context) (Client, error) {
	return &client{}, nil
}

type client struct {
}

func (c *client) ReadConfigFromFile(path string) (config apiv1.Config, err error) {
	log.Debug().Msgf("Reading %v file...", path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		return config, err
	}

	config.SetDefaults()

	return
}
