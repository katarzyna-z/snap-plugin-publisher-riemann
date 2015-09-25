package riemann

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/amir/raidman"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

const (
	PluginName    = "riemann"
	PluginVersion = 2
	PluginType    = plugin.PublisherPluginType
)

// Meta returns the metadata details for the Riemann Publisher Plugin
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(PluginName, PluginVersion, PluginType, []string{plugin.PulseGOBContentType}, []string{plugin.PulseGOBContentType})
}

type riemannPublisher struct{}

// NewRiemannPublisher does something cool
func NewRiemannPublisher() *riemannPublisher {
	return &riemannPublisher{}
}

// GetConfigPolicy returns the config policy for the Riemann Publisher Plugin
func (r *riemannPublisher) GetConfigPolicy() cpolicy.ConfigPolicy {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	// Riemann server to publish event to
	r1, err := cpolicy.NewStringRule("broker", true)
	handleErr(err)
	r1.Description = "Broker in the format of broker-ip:port (ex: 192.168.1.1:5555)"

	config.Add(r1)
	cp.Add([]string{""}, config)
	return *cp
}

// Publish serializes the data and calls publish to send events to Riemann
func (r *riemannPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := log.New()
	//err := r.publish(event, broker)
	//return err
	logger.Println("Riemann Publishing Started")
	var metrics []plugin.PluginMetricType
	switch contentType {
	case plugin.PulseGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.Printf("Error decoding: error=%v content=%v", err, content)
			return err
		}
	default:
		logger.Printf("Error unknown content type '%v'", contentType)
		return errors.New(fmt.Sprintf("Unknown content type '%s'", contentType))
	}
	logger.Printf("publishing %v to %v", metrics, config)
	for _, m := range metrics {
		e := createEvent(m, config)
		if err := r.publish(e, config["broker"].(ctypes.ConfigValueStr).Value); err != nil {
			logger.Println(err)
			return err
		}
	}
	return nil
}

// publish sends events to riemann
func (r *riemannPublisher) publish(event *raidman.Event, broker string) error {
	c, err := raidman.Dial("tcp", broker)
	defer c.Close()
	if err != nil {
		return err
	}
	return c.Send(event)
}

func createEvent(m plugin.PluginMetricType, config map[string]ctypes.ConfigValue) *raidman.Event {
	return &raidman.Event{
		Host:    m.Source(),
		Service: strings.Join(m.Namespace(), "/"),
		Metric:  m.Data(),
	}
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}
