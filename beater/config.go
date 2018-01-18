package beater

import (
	"time"

	"github.com/elastic/apm-server/sourcemap"
	"github.com/elastic/beats/libbeat/common"
)

type Config struct {
	Host               string          `config:"host"`
	MaxUnzippedSize    int64           `config:"max_unzipped_size"`
	MaxHeaderSize      int             `config:"max_header_size"`
	ReadTimeout        time.Duration   `config:"read_timeout"`
	WriteTimeout       time.Duration   `config:"write_timeout"`
	ShutdownTimeout    time.Duration   `config:"shutdown_timeout"`
	SecretToken        string          `config:"secret_token"`
	SSL                *SSLConfig      `config:"ssl"`
	ConcurrentRequests int             `config:"concurrent_requests" validate:"min=1"`
	Frontend           *FrontendConfig `config:"frontend"`
}

type FrontendConfig struct {
	Enabled             *bool          `config:"enabled"`
	RateLimit           int            `config:"rate_limit"`
	AllowOrigins        []string       `config:"allow_origins"`
	LibraryPattern      string         `config:"library_pattern"`
	ExcludeFromGrouping string         `config:"exclude_from_grouping"`
	SourceMapping       *SourceMapping `config:"source_mapping"`
}

type SourceMapping struct {
	Cache *Cache `config:"cache"`
	Index string `config:"index_pattern"`

	esConfig *common.Config
	mapper   sourcemap.Mapper
}

type Cache struct {
	Expiration time.Duration `config:"expiration"`
}

type SSLConfig struct {
	Enabled    *bool  `config:"enabled"`
	PrivateKey string `config:"key"`
	Cert       string `config:"certificate"`
}

func (c *Config) setElasticsearch(esConfig *common.Config) {
	if c != nil && c.Frontend.isEnabled() && c.Frontend.SourceMapping != nil {
		c.Frontend.SourceMapping.esConfig = esConfig
	}
}

func (c *SSLConfig) isEnabled() bool {
	return c != nil && (c.Enabled == nil || *c.Enabled)
}

func (c *FrontendConfig) isEnabled() bool {
	return c != nil && (c.Enabled == nil || *c.Enabled)
}

func (s *SourceMapping) isSetup() bool {
	return s != nil && (s.esConfig != nil)
}

func (c *FrontendConfig) SmapMapper() (sourcemap.Mapper, error) {
	smap := c.SourceMapping
	if !c.isEnabled() || !smap.isSetup() {
		return nil, nil
	}
	if smap.mapper != nil {
		return c.SourceMapping.mapper, nil
	}
	smapConfig := sourcemap.Config{
		CacheExpiration:     smap.Cache.Expiration,
		ElasticsearchConfig: smap.esConfig,
		Index:               smap.Index,
	}
	smapMapper, err := sourcemap.NewSmapMapper(smapConfig)
	if err != nil {
		return nil, err
	}
	c.SourceMapping.mapper = smapMapper
	return c.SourceMapping.mapper, nil
}

func defaultConfig() *Config {
	return &Config{
		Host:               "localhost:8200",
		MaxUnzippedSize:    50 * 1024 * 1024, // 50mb
		MaxHeaderSize:      1 * 1024 * 1024,  // 1mb
		ConcurrentRequests: 40,
		ReadTimeout:        2 * time.Second,
		WriteTimeout:       2 * time.Second,
		ShutdownTimeout:    5 * time.Second,
		SecretToken:        "",
		Frontend: &FrontendConfig{
			Enabled:      new(bool),
			RateLimit:    10,
			AllowOrigins: []string{"*"},
			SourceMapping: &SourceMapping{
				Cache: &Cache{
					Expiration: 5 * time.Minute,
				},
				Index: "apm-*",
			},
			LibraryPattern:      "node_modules|bower_components|~",
			ExcludeFromGrouping: "^/webpack",
		},
	}
}
