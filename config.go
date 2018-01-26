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
	Redis         *RedisConfig      `json:"redis"`
	Gearman       *GearmanConfig    `json:"gearman"`
	Memcached     *MemcachedConfig  `json:"memcached"`
	Mongo         *MongoConfig      `json:"mongo"`
	Hdfs          *HdfsConfig       `json:"hdfs"`
	App           *AppConfig        `json:"app"`
	Jupyter       *JupyterConfig    `json:"jupyter"`
	JobController *JobServerConfig  `json:"jobserver"`
	JobUpdater    *JobUpdaterConfig `json:"jobupdater"`
	Migration     *MigrationConfig  `json:"migration"`
	Influxdb      *InfluxdbConfig   `json:"influxdb"`
	Data          *DataConfig       `json:"data"`

	// the version settings of the current application
	Version string `json:"version"`

	// config for kubernetes service, container application instances like
	// jupyter notebook will be created via this service.
	Kubernetes *KubernetesConfig `json:"kubernetes"`
}

// GetWorkspaceRootDir - Get batch process directory
func (c *Config) GetWorkspaceRootDir() string {
	return c.Data.WorkspaceDir
}

// Return the full path of a workspace directory
func (c *Config) GetWorkspaceDir(workspace *entity.Workspace) string {
	rootDir := c.GetWorkspaceRootDir()
	return filepath.Join(rootDir, workspace.Basename())
}

func (c *Config) FormatWorkspaceBasename(w *entity.Workspace) string {
	return filepath.Join(c.Data.WorkspaceDir, w.Basename())
}

// GetWorkspaceSubpath - this is currently used by PV.  /data will be striped.
// FIXME: the PV related path should be handled.
// FIXME: workspace basename should be saved since we use "batch-{ID}" for all workspaces
func (c *Config) GetWorkspacePVSubpath(w *entity.Workspace) string {
	return filepath.Join(filepath.Base(c.Data.WorkspaceDir), w.Basename())
}

//GetArchiveDir - Get archive directory.
func (c *Config) GetArchiveDir() string {
	return c.Data.ArchiveDir
}

//GetImageDir - Get image directory.
func (c *Config) GetImageDir() string {
	return c.Data.ImageDir
}

//GetThumbnailDir - Get thumbnail directory.
func (c *Config) GetThumbnailDir() string {
	return c.Data.ThumbnailDir
}

//GetModelDir - Get model directory.
func (c *Config) GetModelDir() string {
	return c.Data.ModelDir
}

//GetModelArchiveDir - Get model directory.
func (c *Config) GetModelArchiveDir() string {
	return c.Data.ModelArchiveDir
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
	SetupServiceAddressFromInterface(c.Redis)
	SetupServiceAddressFromInterface(c.Gearman)
	SetupServiceAddressFromInterface(c.Memcached)
}

func Read(path string) (c Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return c, fmt.Errorf("Failed to open the config file: %v\n", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		return c, fmt.Errorf("Failed to load the config file: %v\n", err)
	}
	SetupAddressFromInterface(&c)
	return c, nil
}

func MustRead(path string) Config {
	c, err := Read(path)
	if err != nil {
		panic(err)
	}
	return c
}
