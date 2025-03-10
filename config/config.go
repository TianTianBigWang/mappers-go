/*
Copyright 2022 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kubeedge/mappers-go/pkg/common"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// ErrConfigCert error of certification configuration.
var ErrConfigCert = errors.New("Both certification and private key must be provided")

var defaultConfigFile = "./config.yaml"

// Config is the common mapper configuration.
type Config struct {
	Mqtt       Mqtt       `yaml:"mqtt,omitempty"`
	HttpServer HTTPServer `yaml:"http_server"`
	GrpcServer GRPCServer `yaml:"grpc_server"`
	Common     Common     `yaml:"common"`
	DevInit    DevInit    `yaml:"dev_init"`
}

// Mqtt is the Mqtt configuration.
type Mqtt struct {
	ServerAddress string `yaml:"server,omitempty"`
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Cert          string `yaml:"certification,omitempty"`
	PrivateKey    string `yaml:"privatekey,omitempty"`
}

// MetaServer is the MetaServer configuration.
type MetaServer struct {
	Addr      string `yaml:"addr"`
	Namespace string `json:"namespace"`
}

type HTTPServer struct {
	Host string `yaml:"host"`
}

type GRPCServer struct {
	SocketPath string `yaml:"socket_path"`
}

type Common struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	APIVersion   string `yaml:"api_version"`
	Protocol     string `yaml:"protocol"`
	Address      string `yaml:"address"`
	EdgeCoreSock string `yaml:"edgecore_sock"`
}

type DevInit struct {
	Mode       string     `yaml:"mode"`
	Configmap  string     `yaml:"configmap"`
	MetaServer MetaServer `yaml:"metaserver"`
}

// Parse the configuration file. If failed, return error.
func (c *Config) Parse() error {
	var level klog.Level
	var loglevel string
	var configFile string

	pflag.StringVar(&loglevel, "v", "1", "log level")
	pflag.StringVar(&configFile, "config-file", defaultConfigFile, "Config file name")
	pflag.StringVar(&c.DevInit.MetaServer.Addr, "metaserver-addr", "http://127.0.0.1:10550", "edgecore meta server addr")

	pflag.Parse()
	cf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(cf, c); err != nil {
		return err
	}
	if err = level.Set(loglevel); err != nil {
		return err
	}

	if err = c.parseFlags(); err != nil {
		return err
	}

	switch c.DevInit.Mode {
	case common.DevInitModeConfigmap:
		if readFile, err := ioutil.ReadFile(c.DevInit.Configmap); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			c.DevInit.Configmap = strings.TrimSpace(os.Getenv("DEVICE_PROFILE"))
		} else {
			c.DevInit.Configmap = string(readFile)
		}
		if strings.TrimSpace(c.DevInit.Configmap) == "" {
			return errors.New("can not parse configmap")
		}
	case common.DevInitModeRegister:
	case "": // if mode is nil, use meta server mode
		c.DevInit.Mode = common.DevInitModeMetaServer
		fallthrough
	case common.DevInitModeMetaServer:
		if c.DevInit.MetaServer.Namespace == "" {
			c.DevInit.MetaServer.Namespace = "default"
		}
	default:
		return errors.New("unsupported dev init mode " + c.DevInit.Mode)
	}

	return nil
}

// parseFlags parse flags. Certification and Private key must be provided at the same time.
func (c *Config) parseFlags() error {
	pflag.StringVar(&c.Mqtt.ServerAddress, "mqtt-address", c.Mqtt.ServerAddress, "MQTT broker address")
	pflag.StringVar(&c.Mqtt.Username, "mqtt-username", c.Mqtt.Username, "username")
	pflag.StringVar(&c.Mqtt.Password, "mqtt-password", c.Mqtt.Password, "password")
	pflag.StringVar(&c.Mqtt.Cert, "mqtt-certification", c.Mqtt.Cert, "certification file path")
	pflag.StringVar(&c.Mqtt.PrivateKey, "mqtt-privatekey", c.Mqtt.PrivateKey, "private key file path")
	pflag.Parse()
	if (c.Mqtt.Cert != "" && c.Mqtt.PrivateKey == "") ||
		(c.Mqtt.Cert == "" && c.Mqtt.PrivateKey != "") {
		return ErrConfigCert
	}
	return nil
}
