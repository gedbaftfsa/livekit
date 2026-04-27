package config

import (
	"os"
	"time"

	"github.com/livekit/protocol/logger"
	"gopkg.in/yaml.v3"
)

// Config holds the full server configuration.
type Config struct {
	// Port the server listens on for client connections.
	Port uint32 `yaml:"port"`

	// BindAddresses specifies which network interfaces to bind to.
	BindAddresses []string `yaml:"bind_addresses"`

	// RTC contains real-time communication settings.
	RTC RTCConfig `yaml:"rtc"`

	// Redis connection settings for distributed state.
	Redis RedisConfig `yaml:"redis"`

	// Room contains default room settings.
	Room RoomConfig `yaml:"room"`

	// Logging configuration.
	Logging logger.Config `yaml:"logging"`

	// Keys maps API key to API secret for authentication.
	Keys map[string]string `yaml:"keys"`

	// Region identifies the server's geographic region.
	Region string `yaml:"region"`

	// NodeID is a unique identifier for this server node.
	NodeID string `yaml:"node_id"`
}

// RTCConfig holds WebRTC-specific configuration.
type RTCConfig struct {
	// UDPPort is the port range for UDP media traffic.
	UDPPort uint32 `yaml:"udp_port"`

	// TCPPort is the port for TCP media fallback.
	TCPPort uint32 `yaml:"tcp_port"`

	// ICEServers lists STUN/TURN servers for ICE negotiation.
	ICEServers []ICEServer `yaml:"ice_servers"`

	// UseExternalIP instructs the server to advertise its external IP.
	UseExternalIP bool `yaml:"use_external_ip"`

	// MaxBitrate is the maximum allowed bitrate per participant in bps.
	MaxBitrate uint64 `yaml:"max_bitrate"`
}

// ICEServer represents a STUN or TURN server entry.
type ICEServer struct {
	URLs       []string `yaml:"urls"`
	Username   string   `yaml:"username"`
	Credential string   `yaml:"credential"`
}

// RedisConfig holds connection parameters for Redis.
type RedisConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	UseTLS   bool   `yaml:"use_tls"`
}

// RoomConfig holds default settings applied to newly created rooms.
type RoomConfig struct {
	// EmptyTimeout is how long an empty room persists before being removed.
	EmptyTimeout time.Duration `yaml:"empty_timeout"`

	// MaxParticipants is the default participant cap (0 = unlimited).
	MaxParticipants uint32 `yaml:"max_participants"`

	// EnabledCodecs lists the audio/video codecs the server will negotiate.
	EnabledCodecs []string `yaml:"enabled_codecs"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:          7880,
		BindAddresses: []string{"0.0.0.0"},
		RTC: RTCConfig{
			UDPPort:        7882,
			TCPPort:        7881,
			UseExternalIP:  false,
			MaxBitrate:     3_000_000,
		},
		Room: RoomConfig{
			EmptyTimeout:    5 * time.Minute,
			MaxParticipants: 0,
			EnabledCodecs:   []string{"opus", "vp8", "h264", "vp9", "av1"},
		},
		Keys: map[string]string{},
	}
}

// LoadFromFile reads a YAML configuration file and merges it over defaults.
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
