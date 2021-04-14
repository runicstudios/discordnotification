package handlers

import (
	"testing"
	"time"
	"uwdiscorwb/v1/pkg/config"
)

func Test(t *testing.T)  {
	props, err := config.GetConfig()
	if err != nil {
		t.Errorf("failed to initilize config - %s", err.Error())
	}
	if props.DiscordWebhookUrl == "" {
		t.Errorf("discord webhook not provided")
	}
	emberd := DiscordEmbed{
		Username:  "Unit test: " + time.Now().String(),
		AvatarURL: "",
		Embeds:    nil,
	}
	sendMessageToDiscord(emberd)
}
