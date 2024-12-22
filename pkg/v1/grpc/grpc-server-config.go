package grpc

import "github.com/spf13/viper"

type ServerConfig struct {
	Port              int
	EnableHealthCheck bool
}

func NewServerConfigFromEnv() ServerConfig {
	vi := viper.New()
	vi.AutomaticEnv()

	vi.SetDefault("GRPC_SERVER_PORT", 3000)
	vi.SetDefault("GRPC_SERVER_HEALTH", true)
	return ServerConfig{
		Port:              vi.GetInt("GRPC_SERVER_PORT"),
		EnableHealthCheck: vi.GetBool("GRPC_SERVER_HEALTH"),
	}
}
