package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Configuration is the configuration of services
type Configuration struct {
	Name            string            `yaml:"name" json:"name"`
	Version         string            `yaml:"version" json:"version"`
	Services        []*Service        `yaml:"services,omitempty" json:"services,omitempty"`
	Routes          []*Route          `yaml:"routes,omitempty" json:"routes,omitempty"`
	Consumers       []*Consumer       `yaml:"consumers,omitempty" json:"consumers,omitempty"`
	SSLs            []*SSL            `yaml:"ssls,omitempty" json:"ssls,omitempty"`
	GlobalRules     []*GlobalRule     `yaml:"global_rules,omitempty" json:"global_rules,omitempty"`
	PluginConfigs   []*PluginConfig   `yaml:"plugin_configs,omitempty" json:"plugin_configs,omitempty"`
	ConsumerGroups  []*ConsumerGroup  `yaml:"consumer_groups,omitempty" json:"consumer_groups,omitempty"`
	PluginMetadatas []*PluginMetadata `yaml:"plugin_metadatas,omitempty" json:"plugin_metadatas,omitempty"`
}

// Labels is the APISIX resource labels
type Labels map[string]string

// Vars represents the route match expressions of APISIX.
type Vars [][]StringOrSlice

// Route apisix route object
type Route struct {
	ID string `json:"id" yaml:"id"`

	Name        string `json:"name" yaml:"name"`
	Labels      Labels `json:"labels,omitempty" yaml:"labels,omitempty"`
	Description string `json:"desc,omitempty" yaml:"desc,omitempty"`

	Host            string           `json:"host,omitempty" yaml:"host,omitempty"`
	Hosts           []string         `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Uri             string           `json:"uri,omitempty" yaml:"uri,omitempty"`
	Priority        int              `json:"priority,omitempty" yaml:"priority,omitempty"`
	Timeout         *UpstreamTimeout `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Vars            Vars             `json:"vars,omitempty" yaml:"vars,omitempty"`
	Uris            []string         `json:"uris,omitempty" yaml:"uris,omitempty"`
	Methods         []string         `json:"methods,omitempty" yaml:"methods,omitempty"`
	EnableWebsocket bool             `json:"enable_websocket,omitempty" yaml:"enable_websocket,omitempty"`
	RemoteAddrs     []string         `json:"remote_addrs,omitempty" yaml:"remote_addrs,omitempty"`
	UpstreamId      string           `json:"upstream_id,omitempty" yaml:"upstream_id,omitempty"`
	ServiceID       string           `json:"service_id,omitempty" yaml:"service_id,omitempty"`
	Plugins         Plugins          `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	PluginConfigId  string           `json:"plugin_config_id,omitempty" yaml:"plugin_config_id,omitempty"`
	FilterFunc      string           `json:"filter_func,omitempty" yaml:"filter_func,omitempty"`
}

// Service is the abstraction of a backend service on API gateway.
type Service struct {
	ID string `json:"id" yaml:"id"`

	Name        string `json:"name" yaml:"name"`
	Description string `json:"desc,omitempty" yaml:"desc,omitempty"`
	// Labels are used for resource classification and indexing
	Labels Labels `json:"labels,omitempty" yaml:"labels,omitempty"`
	// HTTP hosts for this service.
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	// Plugin settings on Service level
	Plugins Plugins `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	// Upstream settings for the Service.
	Upstream Upstream `json:"upstream,omitempty" yaml:"upstream,omitempty"`
	// UpstreamId settings for the Service.
	UpstreamId string `json:"upstream_id,omitempty" yaml:"upstream_id,omitempty"`
	// Enables a websocket. Set to false by default.
	EnableWebsocket bool `json:"enable_websocket,omitempty" yaml:"enable_websocket,omitempty"`
}

// Upstream is the definition of the upstream on Service.
type Upstream struct {
	// ID is the upstream name. It should be unique among all upstreams
	// in the same service.
	ID string `json:"id" yaml:"id"`

	Name     string               `json:"name" yaml:"name"`
	Type     string               `json:"type,omitempty" yaml:"type,omitempty"`
	HashOn   string               `json:"hash_on,omitempty" yaml:"hash_on,omitempty"`
	Key      string               `json:"key,omitempty" yaml:"key,omitempty"`
	Checks   *UpstreamHealthCheck `json:"checks,omitempty" yaml:"checks,omitempty"`
	Nodes    UpstreamNodes        `json:"nodes" yaml:"nodes"`
	Scheme   string               `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	Retries  *int                 `json:"retries,omitempty" yaml:"retries,omitempty"`
	Timeout  *UpstreamTimeout     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	TLS      *ClientTLS           `json:"tls,omitempty" yaml:"tls,omitempty"`
	PassHost string               `json:"passhost,omitempty" yaml:"passhostomitempty"`

	// for Service Discovery
	ServiceName   string            `json:"service_name,omitempty" yaml:"service_name,omitempty"`
	DiscoveryType string            `json:"discovery_type,omitempty" yaml:"discovery_type,omitempty"`
	DiscoveryArgs map[string]string `json:"discovery_args,omitempty" yaml:"discovery_args,omitempty"`
}

// UpstreamNode is the node in upstream
type UpstreamNode struct {
	Host   string `json:"host,omitempty" yaml:"host,omitempty"`
	Port   int    `json:"port,omitempty" yaml:"port,omitempty"`
	Weight int    `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// UpstreamNodes is the upstream node list.
type UpstreamNodes []UpstreamNode

// UnmarshalJSON implements json.Unmarshaler interface.
// lua-cjson doesn't distinguish empty array and table,
// and by default empty array will be encoded as '{}'.
// We have to maintain the compatibility.
func (n *UpstreamNodes) UnmarshalJSON(p []byte) error {
	var data []UpstreamNode
	if p[0] == '{' {
		value := map[string]float64{}
		if err := json.Unmarshal(p, &value); err != nil {
			return err
		}
		for k, v := range value {
			node, err := mapKV2Node(k, v)
			if err != nil {
				return err
			}
			data = append(data, *node)
		}
		*n = data
		return nil
	}
	if err := json.Unmarshal(p, &data); err != nil {
		return err
	}
	*n = data
	return nil
}

// UpstreamHealthCheck defines the active and/or passive health check for an Upstream,
// with the upstream health check feature, pods can be kicked out or joined in quickly,
// if the feedback of Kubernetes liveness/readiness probe is long.
type UpstreamHealthCheck struct {
	Active  *UpstreamActiveHealthCheck  `json:"active" yaml:"active"`
	Passive *UpstreamPassiveHealthCheck `json:"passive,omitempty" yaml:"passive,omitempty"`
}

// UpstreamActiveHealthCheck defines the active kind of upstream health check.
type UpstreamActiveHealthCheck struct {
	Type               string                             `json:"type,omitempty" yaml:"type,omitempty"`
	Timeout            int                                `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Concurrency        int                                `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	Host               string                             `json:"host,omitempty" yaml:"host,omitempty"`
	Port               int32                              `json:"port,omitempty" yaml:"port,omitempty"`
	HTTPPath           string                             `json:"http_path,omitempty" yaml:"http_path,omitempty"`
	HTTPSVerifyCert    bool                               `json:"https_verify_certificate,omitempty" yaml:"https_verify_certificate,omitempty"`
	HTTPRequestHeaders []string                           `json:"req_headers,omitempty" yaml:"req_headers,omitempty"`
	Healthy            UpstreamActiveHealthCheckHealthy   `json:"healthy,omitempty" yaml:"healthy,omitempty"`
	Unhealthy          UpstreamActiveHealthCheckUnhealthy `json:"unhealthy,omitempty" yaml:"unhealthy,omitempty"`
}

// UpstreamPassiveHealthCheck defines the passive kind of upstream health check.
type UpstreamPassiveHealthCheck struct {
	Type      string                              `json:"type,omitempty" yaml:"type,omitempty"`
	Healthy   UpstreamPassiveHealthCheckHealthy   `json:"healthy,omitempty" yaml:"healthy,omitempty"`
	Unhealthy UpstreamPassiveHealthCheckUnhealthy `json:"unhealthy,omitempty" yaml:"unhealthy,omitempty"`
}

// UpstreamActiveHealthCheckHealthy defines the conditions to judge whether
// an upstream node is healthy with the active manner.
type UpstreamActiveHealthCheckHealthy struct {
	UpstreamPassiveHealthCheckHealthy `json:",inline" yaml:",inline"`

	Interval int `json:"interval,omitempty" yaml:"interval,omitempty"`
}

// UpstreamPassiveHealthCheckHealthy defines the conditions to judge whether
// an upstream node is healthy with the passive manner.
type UpstreamPassiveHealthCheckHealthy struct {
	HTTPStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	Successes    int   `json:"successes,omitempty" yaml:"successes,omitempty"`
}

// UpstreamActiveHealthCheckUnhealthy defines the conditions to judge whether
// an upstream node is unhealthy with the active manager.
type UpstreamActiveHealthCheckUnhealthy struct {
	UpstreamPassiveHealthCheckUnhealthy `json:",inline" yaml:",inline"`

	Interval int `json:"interval,omitempty" yaml:"interval,omitempty"`
}

// UpstreamPassiveHealthCheckUnhealthy defines the conditions to judge whether
// an upstream node is unhealthy with the passive manager.
type UpstreamPassiveHealthCheckUnhealthy struct {
	HTTPStatuses []int `json:"http_statuses,omitempty" yaml:"http_statuses,omitempty"`
	HTTPFailures int   `json:"http_failures,omitempty" yaml:"http_failures,omitempty"`
	TCPFailures  int   `json:"tcp_failures,omitempty" yaml:"tcp_failures,omitempty"`
	Timeouts     int   `json:"timeouts,omitempty" yaml:"timeouts,omitempty"`
}

// ClientTLS is tls cert and key use in mTLS
type ClientTLS struct {
	Cert string `json:"client_cert,omitempty" yaml:"client_cert,omitempty"`
	Key  string `json:"client_key,omitempty" yaml:"client_key,omitempty"`
}

// UpstreamTimeout represents the timeout settings on Upstream.
type UpstreamTimeout struct {
	// Connect is the connect timeout
	Connect int `json:"connect" yaml:"connect"`
	// Send is the send timeout
	Send int `json:"send" yaml:"send"`
	// Read is the read timeout
	Read int `json:"read" yaml:"read"`
}

func mapKV2Node(key string, val float64) (*UpstreamNode, error) {
	hp := strings.Split(key, ":")
	host := hp[0]
	//  according to APISIX upstream nodes policy, port is required
	port := "80"

	if len(hp) > 2 {
		return nil, errors.New("invalid upstream node")
	} else if len(hp) == 2 {
		port = hp[1]
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("parse port to int fail: %s", err.Error())
	}

	node := &UpstreamNode{
		Host:   host,
		Port:   portInt,
		Weight: int(val),
	}

	return node, nil
}

type Plugin map[string]interface{}
type Plugins map[string]Plugin

func (p *Plugins) DeepCopyInto(out *Plugins) {
	b, _ := json.Marshal(&p)
	_ = json.Unmarshal(b, out)
}

func (p *Plugins) DeepCopy() *Plugins {
	if p == nil {
		return nil
	}
	out := new(Plugins)
	p.DeepCopyInto(out)
	return out
}

func (p *Plugins) UnmarshalJSON(cont []byte) error {
	var plugins map[string]Plugin
	err := json.Unmarshal(cont, &plugins)
	if err != nil {
		return err
	}

	if p == nil || *p == nil {
		*p = make(Plugins)
	}
	for name, config := range plugins {
		(*p)[name] = GetPluginDefaultValues(name, config)
	}

	return nil
}

// StringOrSlice represents a string or a string slice.
// TODO Do not use interface{} to avoid the reflection overheads.
type StringOrSlice struct {
	StrVal   string   `json:"-"`
	SliceVal []string `json:"-"`
}

func (s *StringOrSlice) MarshalJSON() ([]byte, error) {
	var (
		p   []byte
		err error
	)
	if s.SliceVal != nil {
		p, err = json.Marshal(s.SliceVal)
	} else {
		p, err = json.Marshal(s.StrVal)
	}
	return p, err
}

func (s *StringOrSlice) UnmarshalJSON(p []byte) error {
	var err error

	if len(p) == 0 {
		return errors.New("empty object")
	}
	if p[0] == '[' {
		err = json.Unmarshal(p, &s.SliceVal)
	} else {
		err = json.Unmarshal(p, &s.StrVal)
	}
	return err
}

// Consumer represents the consumer object in APISIX.
type Consumer struct {
	Username string `json:"username" yaml:"username"`
	Desc     string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Labels   Labels `json:"labels,omitempty" yaml:"labels,omitempty"`

	Plugins Plugins `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	GroupID string  `json:"group_id,omitempty" yaml:"group_id,omitempty"`
}

// SSL represents the ssl object in APISIX.
type SSL struct {
	ID     string `json:"id" yaml:"id"`
	Labels Labels `json:"labels,omitempty" yaml:"labels,omitempty"`

	SNIs []string `json:"snis" yaml:"snis"`
	Cert string   `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key  string   `json:"key,omitempty" yaml:"key,omitempty"`
}

// GlobalRule represents the global_rule object in APISIX.
type GlobalRule struct {
	ID      string  `json:"id" yaml:"id"`
	Plugins Plugins `json:"plugins" yaml:"plugins"`
}

// PluginConfig apisix plugin object
type PluginConfig struct {
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Desc   string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Labels Labels `json:"labels,omitempty" yaml:"labels,omitempty"`

	Plugins Plugins `json:"plugins" yaml:"plugins"`
}

// ConsumerGroup apisix consumer group object
type ConsumerGroup struct {
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Desc   string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Labels Labels `json:"labels,omitempty" yaml:"labels,omitempty"`

	Plugins Plugins `json:"plugins" yaml:"plugins"`
}

const (
	UpstreamPassHost = "host"
)

type PluginMetadata struct {
	ID     string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Config map[string]interface{} `json:",inline" yaml:",inline"`
}

func (s *PluginMetadata) MarshalJSON() ([]byte, error) {
	if s.Config == nil {
		s.Config = make(map[string]interface{})
	}

	s.Config["id"] = s.ID

	p, err := json.Marshal(s.Config)

	delete(s.Config, "id")

	return p, err
}

func (s *PluginMetadata) UnmarshalJSON(p []byte) error {
	var (
		config map[string]interface{}
		err    error
	)

	err = json.Unmarshal(p, &config)
	if err != nil {
		return err
	}

	id, ok := config["id"]
	if ok {
		if reflect.TypeOf(id).Kind() != reflect.String {
			return errors.New("plugin metadata id is not a string, input: " + string(p))
		}
		s.ID = fmt.Sprintf("%v", id)
		delete(config, "id")
	}

	s.Config = config

	return nil
}
