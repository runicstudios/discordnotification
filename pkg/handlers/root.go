package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/jeremywohl/flatten"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"uwdiscorwb/v1/pkg/config"
	"uwdiscorwb/v1/pkg/log"
	"uwdiscorwb/v1/pkg/helpers"
	"uwdiscorwb/v1/pkg/types"
)

var properties, _ = config.GetConfig()

type Embed struct {
	Description string `json:"description"`
	Color int `json:"color"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

// discord embed object to show formatted messages to users
type DiscordEmbed struct {
	Username string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Embeds []Embed `json:"embeds"`
}

// voipms callback message
type voipmsMessage struct {
	To string  `json:"to" validate:"required"`
	From string `json:"from" validate:"required"`
	Message string `json:"message" validate:"required"`
	Id string `json:"id" validate:"required"`
	Timestamp string `json:"timestamp" validate:"required"`
}

// HealthCheck handler for health route, mostly used for debugging
// gives a static response
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, types.RespMessage{
		Success: true,
		Message: "Message server is running and ready to access connections !!",
		Data:    nil,
	})
}

// FlowrouteCallbackHandler handle incoming call from the flowroute and forward
// the helpers to discord channel as text
func FlowrouteCallbackHandler(c echo.Context) error {
	buf, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		 log.Error("failed to read helpers body, bad helpers")
		 return c.JSON(http.StatusBadRequest, "Bad helpers")
	}
	msg, err := formatMessage(string(buf), false)
	if err != nil {
		msg = []Field{
			{
				Name: "Message",
				Value: fmt.Sprintf("%+v", string(buf)),
			},
		}
	}
	embed := DiscordEmbed{
		Username:  "Flowroute Message Received!",
		AvatarURL: "https://imgur.com/a/Cl3zspb",
		Embeds:    []Embed{
			{
				Description: "",
				Color: 65280,
				Fields: msg,
			},
		},
	}
	err, status := sendMessageToDiscord(embed)
	if err != nil {
		return c.JSON(status, types.RespMessage{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}
	return c.JSON(http.StatusOK, "OK")
}

//voipmsCallbackHandler handles webhooks calls for voipms,
// {URL}?to={TO}&from={FROM}&message={MESSAGE}&id={ID}&date={TIMESTAMP}
func VoipmsCallbackHandler(c echo.Context) error {
	msg := voipmsMessage{
		To:        c.QueryParam("to"),
		From:      c.QueryParam("from"),
		Message:   c.QueryParam("message"),
		Id:        c.QueryParam("id"),
		Timestamp: c.QueryParam("date"),
	}
	badReq := helpers.Validate(msg)
	if badReq != nil {
		return c.JSON(http.StatusBadRequest, badReq)
	}
	buf, _ := json.Marshal(msg)
	fields, err := formatMessage(string(buf), true)
	embed := DiscordEmbed{
		Username:  "Voipms Message Received!",
		AvatarURL: "https://imgur.com/a/Cl3zspb",
		Embeds:    []Embed{
			{
				Description: "",
				Color: 65280,
				Fields: fields,
			},
		},
	}
	err, status := sendMessageToDiscord(embed)
	if err != nil {
		return c.JSON(status,  types.RespMessage{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}
	return c.JSON(http.StatusOK, "OK")
}

// sendMessageToDiscord will do a http to webhook url of discord server
func sendMessageToDiscord(message DiscordEmbed) (error, int) {
	client := helpers.NewClient()
	log.Debug("received payload - ", message)
	opts := helpers.Options{
		Url:     properties.DiscordWebhookUrl,
		Method:  http.MethodPost,
		Body:    message,
	}
	req, err := client.NewRequest(opts)
	if err != nil {
		log.Error(fmt.Sprintf("failed to build up helpers body for url %s - err : %s", properties.DiscordWebhookUrl, err.Error()))
		return err, http.StatusInternalServerError
	}
	resp, err := req.Send()
	if err != nil {
		log.Error(fmt.Sprintf("helpers sent to discord server but failed with error : %v", err))
		return err, http.StatusInternalServerError
	}
	if resp.GetStatusCode() != http.StatusOK && resp.GetStatusCode() != http.StatusNoContent && resp.GetStatusCode() != http.StatusAccepted {
		body, _ := resp.GetBody()
		msg := fmt.Sprintf("failed discord helpers with status code - %d - response body - %s", resp.GetStatusCode(), string(body))
		log.Error(msg)
		return fmt.Errorf(msg), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

// formatMessage will create a readable message from json string
func formatMessage(msg string, timestamp bool) ([]Field, error) {
	var resp map[string]interface{}
	parsed, _ := flatten.FlattenString(msg, "", flatten.DotStyle)
	err := json.Unmarshal([]byte(parsed), &resp)
	if err != nil {
		return nil, err
	}
	incr := 0
	if timestamp {
		incr += 1
	}
	var fileds []Field
	var initialkeys = make([]string, incr + 4)
	var keys []string
	for k, _ := range resp {
		switch strings.ToLower(k) {
		case "body", "data.attributes.body":
			initialkeys[incr + 3] = k
		case "to", "data.attributes.to":
			initialkeys[incr + 2] = k
		case "from", "data.attributes.from":
			initialkeys[incr + 1] = k
		case "id", "data.id":
			initialkeys[0] = k
		case "timestamp":
			initialkeys[1] = k
		}
	}
	for k, _ := range resp {
		if !contains(initialkeys, k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	keys = append(initialkeys, keys...)
	for _, k := range keys {
		if k == "" {
			continue
		}
		switch resp[k].(type) {
		case string:
			fileds = append(fileds, Field{
				Name:  strings.ToUpper(k),
				Value: resp[k].(string),
			})
		case int:
			fileds = append(fileds, Field{
				Name:  strings.ToUpper(k),
				Value: fmt.Sprintf("%d", resp[k].(int)),
			})
		case interface{}:
			str, _ := json.Marshal(resp[k])
			flds, err := formatMessage(string(str), timestamp)
			if err != nil {
				fileds = append(fileds, Field{
					Name:  strings.ToUpper(k),
					Value: fmt.Sprintf("%v", resp[k]),
				})
			} else {
				fileds = append(fileds, flds...)
			}
		default:
			fileds = append(fileds, Field{
				Name:  strings.ToUpper(k),
				Value: fmt.Sprintf("%v", resp[k]),
			})
		}
	}
	return fileds, nil
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}