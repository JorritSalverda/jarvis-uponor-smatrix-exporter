package bigquery

import (
	"context"
	"fmt"
	"time"

	googlebigquery "cloud.google.com/go/bigquery"
	contractsv1 "github.com/JorritSalverda/jarvis-contracts-golang/contracts/v1"
	"github.com/rs/zerolog/log"
)

// Client is the interface for connecting to bigquery
type Client interface {
	CheckIfDatasetExists(dataset string) (exists bool)
	CheckIfTableExists(dataset, table string) (exists bool)
	CreateTable(dataset, table string, typeForSchema interface{}, partitionField string, waitReady bool) (err error)
	UpdateTableSchema(dataset, table string, typeForSchema interface{}) (err error)
	DeleteTable(dataset, table string) (err error)
	InsertMeasurement(dataset, table string, measurement contractsv1.Measurement) (err error)
	InitBigqueryTable(dataset, table string) (err error)
}

// NewClient returns new bigquery.Client
func NewClient(projectID string, enable bool) (Client, error) {

	ctx := context.Background()

	bigqueryClient, err := googlebigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &client{
		projectID: projectID,
		client:    bigqueryClient,
		enable:    enable,
	}, nil
}

type client struct {
	projectID string
	client    *googlebigquery.Client
	enable    bool
}

func (c *client) CheckIfDatasetExists(dataset string) (exists bool) {

	if !c.enable {
		return false
	}

	ds := c.client.Dataset(dataset)

	md, err := ds.Metadata(context.Background())

	log.Error().Err(err).Msgf("Error retrieving metadata for dataset %v", dataset)

	return md != nil
}

func (c *client) CheckIfTableExists(dataset, table string) (exists bool) {

	if !c.enable {
		return false
	}

	tbl := c.client.Dataset(dataset).Table(table)

	md, _ := tbl.Metadata(context.Background())

	// log.Error().Err(err).Msgf("Error retrieving metadata for table %v", table)

	return md != nil
}

func (c *client) CreateTable(dataset, table string, typeForSchema interface{}, partitionField string, waitReady bool) (err error) {

	if !c.enable {
		return nil
	}

	tbl := c.client.Dataset(dataset).Table(table)

	// infer the schema of the type
	schema, err := googlebigquery.InferSchema(typeForSchema)
	if err != nil {
		return err
	}

	tableMetadata := &googlebigquery.TableMetadata{
		Schema: schema,
	}

	// if partitionField is set use it for time partitioning
	if partitionField != "" {
		tableMetadata.TimePartitioning = &googlebigquery.TimePartitioning{
			Field: partitionField,
		}
	}

	// create the table
	err = tbl.Create(context.Background(), tableMetadata)
	if err != nil {
		return err
	}

	if waitReady {
		for {
			if c.CheckIfTableExists(dataset, table) {
				break
			}
			time.Sleep(time.Second)
		}
	}

	return nil
}

func (c *client) UpdateTableSchema(dataset, table string, typeForSchema interface{}) (err error) {

	if !c.enable {
		return nil
	}

	tbl := c.client.Dataset(dataset).Table(table)

	// infer the schema of the type
	schema, err := googlebigquery.InferSchema(typeForSchema)
	if err != nil {
		return err
	}

	meta, err := tbl.Metadata(context.Background())
	if err != nil {
		return err
	}

	update := googlebigquery.TableMetadataToUpdate{
		Schema: schema,
	}
	if _, err := tbl.Update(context.Background(), update, meta.ETag); err != nil {
		return err
	}

	return nil
}

func (c *client) DeleteTable(dataset, table string) (err error) {

	if !c.enable {
		return nil
	}

	tbl := c.client.Dataset(dataset).Table(table)

	// delete the table
	err = tbl.Delete(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (c *client) InsertMeasurement(dataset, table string, measurement contractsv1.Measurement) (err error) {

	if !c.enable {
		return nil
	}

	tbl := c.client.Dataset(dataset).Table(table)

	u := tbl.Uploader()

	if err := u.Put(context.Background(), measurement); err != nil {
		return err
	}

	return nil
}

func (c *client) InitBigqueryTable(dataset, table string) (err error) {

	log.Debug().Msgf("Checking if table %v.%v.%v exists...", c.projectID, dataset, table)
	tableExist := c.CheckIfTableExists(dataset, table)

	if !tableExist {
		log.Debug().Msgf("Creating table %v.%v.%v...", c.projectID, dataset, table)
		err := c.CreateTable(dataset, table, contractsv1.Measurement{}, "MeasuredAtTime", true)
		if err != nil {
			return fmt.Errorf("Failed creating bigquery table: %w", err)
		}
	} else {
		log.Debug().Msgf("Trying to update table %v.%v.%v schema...", c.projectID, dataset, table)
		err := c.UpdateTableSchema(dataset, table, contractsv1.Measurement{})
		if err != nil {
			return fmt.Errorf("Failed updating bigquery table schema: %w", err)
		}
	}

	return nil
}
