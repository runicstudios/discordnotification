package types

// server properties for optimising the server working
// it also defines the dynamic parts of the server
type Properties struct {
	Port int `mapstructure:"port" json:"port"`
	ServerLog bool `mapstructure:"server_log" json:"server_log"`
	DiscordWebhookUrl string `mapstructure:"discord_webhook_url" json:"discord_webhook_url"`
	SecureServer bool `mapstructure:"secure_server" json:"secure_server"`
}
