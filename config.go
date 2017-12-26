package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/utils/netutils"
)

type ServiceConfig interface {
	SetHost(host string)
	SetPort(port int32)
	GetInterface() string
	Unresolved() bool
	GetPublic() ServiceConfig
	LoadDefaults() error
}

type Config struct {
	Redis         *RedisConfig     `json:"redis"`
	Gearman       *GearmanConfig   `json:"gearman"`
	Memcached     *MemcachedConfig `json:"memcached"`
	Mongo         *MongoConfig     `json:"mongo"`
	Hdfs          *HdfsConfig      `json:"hdfs"`
	App           *AppConfig       `json:"app"`
	Jupyter       *JupyterConfig   `json:"jupyter"`
	JobController *JobServerConfig `json:"jobserver"`
	Influxdb      *InfluxdbConfig  `json:"influxdb"`
	DataDir       string           `json:"dataDir"`
	Data          *DataConfig      `json:"data"`
	Version       string           `json:"version"`

	Kubernetes *KubernetesConfig `json:"kubernetes"`
}

//GetWorkspaceRootDir - Get batch process directory
func (c *Config) GetWorkspaceRootDir() string {
	return filepath.Join(c.DataDir, c.Data.BatchDir)
}

func (c *Config) GetWorkspaceDir(workspace *entity.Workspace) string {
	rootDir := c.GetWorkspaceRootDir()
	return filepath.Join(rootDir, fmt.Sprintf("batch-%s", workspace.ID.Hex()))
}

func (c *Config) FormatWorkspaceBasename(workspace *entity.Workspace) string {
	return filepath.Join(c.Data.BatchDir, fmt.Sprintf("batch-%s", workspace.ID.Hex()))
}

//GetArchiveDir - Get archive directory.
func (c *Config) GetArchiveDir() string {
	return filepath.Join(c.DataDir, c.Data.ArchiveDir)
}

//GetImageDir - Get image directory.
func (c *Config) GetImageDir() string {
	return filepath.Join(c.DataDir, c.Data.ImageDir)
}

//GetThumbnailDir - Get thumbnail directory.
func (c *Config) GetThumbnailDir() string {
	return filepath.Join(c.DataDir, c.Data.ThumbnailDir)
}

//GetModelDir - Get model directory.
func (c *Config) GetModelDir() string {
	return filepath.Join(c.DataDir, c.Data.ModelDir)
}

//GetModelArchiveDir - Get model directory.
func (c *Config) GetModelArchiveDir() string {
	return filepath.Join(c.DataDir, c.Data.ModelArchiveDir)
}

func SetupServiceAddressFromInterface(c ServiceConfig) {
	if reflect.ValueOf(c).IsNil() {
		return
	}
	if c.Unresolved() {
		var inf = c.GetInterface()
		if inf != "" {
			var ip net.IP = netutils.MustFindInterfaceGlobalUnicastIp(inf)
			if ip != nil {
				name := reflect.TypeOf(c).String()
				log.Printf("%s: discovered IP %s on Interface %s\n", name, ip, inf)
				c.SetHost(ip.String())
			}
		}
	}
	c.LoadDefaults()

	if pc := c.GetPublic(); pc != nil {
		SetupServiceAddressFromInterface(pc)
	}
}

// SetupAddressFromInterface will scan the network interface setting if "interface" field
// is defined.
func SetupAddressFromInterface(c *Config) {
	log.Println("Resolving service configurations...")
	SetupServiceAddressFromInterface(c.Redis)
	SetupServiceAddressFromInterface(c.Gearman)
	SetupServiceAddressFromInterface(c.Memcached)
}

func Read(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %v\n", err)
	}
	decoder := json.NewDecoder(file)
	c := Config{}
	if err := decoder.Decode(&c); err != nil {
		log.Fatalf("Failed to load config file: %v\n", err)
	}
	SetupAddressFromInterface(&c)
	return c
}
