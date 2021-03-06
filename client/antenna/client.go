package antenna

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	contractsv1 "github.com/JorritSalverda/jarvis-contracts-golang/contracts/v1"
	apiv1 "github.com/JorritSalverda/jarvis-uponor-smatrix-exporter/api/v1"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/jacobsa/go-serial/serial"
	"github.com/rs/zerolog/log"
)

// Client is the interface for connecting to a websocket device via ethernet
type Client interface {
	GetMeasurement(config apiv1.Config) (measurement contractsv1.Measurement, err error)
	GetSample(config apiv1.Config, sampleConfig apiv1.ConfigSample) (sample contractsv1.Sample, err error)
}

// NewClient returns new websocket.Client
func NewClient(antennaUSBDevicePath string, waitGroup *sync.WaitGroup, done chan struct{}) (Client, error) {
	if antennaUSBDevicePath == "" {
		return nil, fmt.Errorf("Please set the usb device path for the antenna")
	}

	return &client{
		antennaUSBDevicePath: antennaUSBDevicePath,
		waitGroup:            waitGroup,
		lastReceivedMessage:  time.Now().UTC(),
		done:                 done,
	}, nil
}

type client struct {
	antennaUSBDevicePath string
	waitGroup            *sync.WaitGroup

	f                   io.ReadWriteCloser
	in                  *bufio.Reader
	responseChannel     chan []byte
	lastReceivedMessage time.Time
	done                chan struct{}
	teardown            bool
}

func (c *client) GetMeasurement(config apiv1.Config) (measurement contractsv1.Measurement, err error) {

	log.Info().Msg("Starting retrieval of measurement...")

	c.openSerialPort()
	defer c.closeSerialPort()
	go c.keepSerialPortAlive()

	c.responseChannel = make(chan []byte)
	go c.receiveResponse()

	<-c.done
	log.Info().Msg("Received done signal, tearing down serial port listener")
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()
	c.teardown = true

	// measurement = contractsv1.Measurement{
	// 	ID:             uuid.New().String(),
	// 	Source:         "jarvis-uponor-smatrix-exporter",
	// 	Location:       config.Location,
	// 	Samples:        []*contractsv1.Sample{},
	// 	MeasuredAtTime: time.Now().UTC(),
	// }

	// for _, sc := range config.SampleConfigs {
	// 	sample, sampleErr := c.GetSample(config, sc)
	// 	if sampleErr != nil {
	// 		return measurement, sampleErr
	// 	}
	// 	measurement.Samples = append(measurement.Samples, &sample)
	// }

	return
}

func (c *client) GetSample(config apiv1.Config, sampleConfig apiv1.ConfigSample) (sample contractsv1.Sample, err error) {

	// init sample from config
	sample = contractsv1.Sample{
		EntityType: sampleConfig.EntityType,
		EntityName: sampleConfig.EntityName,
		SampleType: sampleConfig.SampleType,
		SampleName: sampleConfig.SampleName,
		MetricType: sampleConfig.MetricType,
	}

	// convert sample to float and correct
	// sample.Value = value * sampleConfig.ValueMultiplier

	return
}

func (c *client) openSerialPort() {
	options := serial.OpenOptions{
		PortName:               c.antennaUSBDevicePath,
		BaudRate:               16550,
		DataBits:               8,
		StopBits:               1,
		MinimumReadSize:        0,
		InterCharacterTimeout:  2000,
		ParityMode:             serial.PARITY_NONE,
		Rs485Enable:            false,
		Rs485RtsHighDuringSend: false,
		Rs485RtsHighAfterSend:  false,
	}

	f, err := serial.Open(options)
	if err != nil {
		log.Fatal().Err(err).Interface("options", options).Msg("Failed opening serial device")
	}

	c.f = f
	c.in = bufio.NewReader(f)
}

func (c *client) closeSerialPort() {
	c.f.Close()

	time.Sleep(5 * time.Second)
}

func (c *client) resetSerialPort() {

	// wait for any previous serial port reset to finish before continuing
	c.waitGroup.Wait()

	// lock area during reset
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	// perform the reset
	c.closeSerialPort()
	c.openSerialPort()
}

func (c *client) keepSerialPortAlive() {
	for {
		time.Sleep(time.Duration(foundation.ApplyJitter(120)) * time.Second)

		if time.Since(c.lastReceivedMessage).Minutes() > 2 {
			log.Info().Msg("Received last message more than 2 minutes ago, resetting serial port...")
			c.resetSerialPort()
		}
	}
}

func (c *client) receiveResponse() (err error) {
	// execute commands and read from serial port
	for {
		// wait for serial port reset to finish before continuing
		c.waitGroup.Wait()

		// read from serial port
		buf, isPrefix, err := c.in.ReadLine()
		if c.teardown {
			log.Info().Msg("Completing teardown of serial port listener")
			return nil
		}

		if err != nil {
			if err != io.EOF {
				log.Warn().Err(err).Msg("Error reading from serial port, resetting port...")
				c.resetSerialPort()
			}
		} else if isPrefix {
			log.Warn().Str("_msg", string(buf)).Msgf("Message is too long for buffer and split over multiple lines")
		} else {

			c.lastReceivedMessage = time.Now().UTC()
			// c.responseChannel <- buf

			rawmsg := string(buf)
			length := len(rawmsg)

			// make sure no obvious errors in getting the data....
			if length > 40 &&
				!strings.Contains(rawmsg, "_ENC") &&
				!strings.Contains(rawmsg, "_BAD") &&
				!strings.Contains(rawmsg, "BAD") &&
				!strings.Contains(rawmsg, "ERR") {

				isValidMessage, err := regexp.MatchString(`^\d{3} ( I| W|RQ|RP) --- (--:------|\d{2}:\d{6}) (--:------ |\d{2}:\d{6} ){2}[0-9a-fA-F]{4} \d{3}`, rawmsg)
				if err != nil || !isValidMessage {
					log.Info().Msgf("read: %v", rawmsg)
				} else {
					log.Debug().Msgf("evohome: %v / %v", rawmsg, buf)
				}
			} else {
				log.Info().Msgf("read: %v / %v", rawmsg, buf)
			}
		}
	}
}
