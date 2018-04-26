package config

import (
	//	"io/ioutil"
	//	"plugin"
	"net/url"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// DefaultConfig section in yml file
type DefaultConfig struct {
	PluginDirectory string
	NumberOfWorkers int
	MaxJobQueueLen  int
}

// EtcdNotifierConfig section in yml file
type EtcdNotifierConfig struct {
	EtcdServers      []string
	WatchPath        string
	MsgQueueLockTime int
	MsgIndexString   string
	ReadLockString   string
	MasterElection   bool
}

// PluginConfig section in yml file
type PluginConfig struct {
	Handlers map[string]interface{}
}

// Config Object
type Config struct {
	FileName        string
	EtcdServersUrls []string
	EtcdServers     []string
	DefaultCfg      *DefaultConfig
	EtcdNotifierCfg *EtcdNotifierConfig
	PluginCfg       *PluginConfig
	PluginNames     []string
}

// ReadConfig reads the configuration file
func (c *Config) ReadConfig() error {
	c.DefaultCfg = &DefaultConfig{
		PluginDirectory: viper.GetString("compilation.plugin_directory"),
		NumberOfWorkers: viper.GetInt("compilation.number_of_workers"),
		MaxJobQueueLen:  viper.GetInt("compilation.max_job_queue_len"),
	}

	c.EtcdNotifierCfg = &EtcdNotifierConfig{
		EtcdServers:      viper.GetStringSlice("etcd.endpoints"),
		WatchPath:        viper.GetString("etcd.path"),
		MsgQueueLockTime: viper.GetInt("compilation.msg_queue_lock_time"),
		MsgIndexString:   viper.GetString("compilation.msg_index_string"),
		ReadLockString:   viper.GetString("compilation.read_lock_string"),
		MasterElection:   viper.GetBool("compilation.master_election"),
	}

	c.PluginCfg = &PluginConfig{
		Handlers: viper.GetStringMap("plugin.handlers"),
	}

	c.EtcdServersUrls = c.EtcdNotifierCfg.EtcdServers
	for _, svr := range c.EtcdServersUrls {
		u, err := url.Parse(svr)
		if err != nil {
			log.Fatal(err)
		}
		c.EtcdServers = append(c.EtcdServers, u.Hostname()+":"+u.Port())
	}

	log.Println("Plugin Directory:", c.DefaultCfg.PluginDirectory)
	log.Println("Number of Workers:", c.DefaultCfg.NumberOfWorkers)
	log.Println("Maximum Job Queue Len:", c.DefaultCfg.MaxJobQueueLen)

	log.Println("ETCD Notifier Servers List:", c.EtcdServersUrls)
	log.Println("ETCD Notifier Servers List:", c.EtcdServers)
	log.Println("ETCD Notifier WatchPath :", c.EtcdNotifierCfg.WatchPath)
	log.Println("ETCD Notifier MsgQueueLockTime:", c.EtcdNotifierCfg.MsgQueueLockTime)
	log.Println("ETCD Notifier MsgIndexString:", c.EtcdNotifierCfg.MsgIndexString)
	log.Println("ETCD Notifier ReadLockString:", c.EtcdNotifierCfg.ReadLockString)
	log.Println("ETCD Notifier MasterElection:", c.EtcdNotifierCfg.MasterElection)

	log.Println("Plugin Handlers:", c.PluginCfg.Handlers)

	return nil
}

// NewConfig creates the Config object
func NewConfig(configFile string) (*Config, error) {
	conf := &Config{
		FileName: configFile,
	}
	err := conf.ReadConfig()
	if err != nil {
		return nil, err
	}
	log.Println("Config file Read")
	return conf, nil
}
