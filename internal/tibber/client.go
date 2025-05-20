package tibber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type WebsocketConfig struct {
	Token string
	Host  string
	Path  string
	Id    string
}

type WebsocketClient struct {
	Config *WebsocketConfig
	Data   chan Measurement
}

type Client struct {
	WebsocketClient *WebsocketClient
	Wg              *sync.WaitGroup
}

type Message struct {
	Type    string          `json:"type"`
	Id      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Measurement struct {
	Timestamp              time.Time `json:"timestamp"`
	Power                  float64   `json:"power"`
	PowerProduction        float64   `json:"powerProduction"`
	MinPower               float64   `json:"minPower"`
	AveragePower           float64   `json:"averagePower"`
	MaxPower               float64   `json:"maxPower"`
	MaxPowerProduction     float64   `json:"maxPowerProduction"`
	AccumulatedConsumption float64   `json:"accumulatedConsumption"`
	AccumulatedProduction  float64   `json:"accumulatedProduction"`
	LastMeterConsumption   float64   `json:"lastMeterConsumption"`
	LastMeterProduction    float64   `json:"lastMeterProduction"`
	CurrentL1              *float64  `json:"currentL1,omitempty"`
	CurrentL2              *float64  `json:"currentL2,omitempty"`
	CurrentL3              *float64  `json:"currentL3,omitempty"`
	VoltagePhase1          *float64  `json:"voltagePhase1,omitempty"`
	VoltagePhase2          *float64  `json:"voltagePhase2,omitempty"`
	VoltagePhase3          *float64  `json:"voltagePhase3,omitempty"`
}

func NewWebsocketConfig(config *WebsocketConfig) *WebsocketConfig {
	return &WebsocketConfig{
		Token: config.Token,
		Host:  config.Host,
		Path:  config.Path,
		Id:    config.Id,
	}
}

func NewClient(token, houseId string) *Client {
	client := &Client{
		WebsocketClient: &WebsocketClient{
			Config: NewWebsocketConfig(&WebsocketConfig{
				Token: token,
				Host:  "websocket-api.tibber.com",
				Path:  "/v1-beta/gql/subscriptions",
				Id:    houseId,
			}),
			Data: make(chan Measurement),
		},
		Wg: &sync.WaitGroup{},
	}

	if err := client.VerifyAccess(); err != nil {
		log.Printf("Warning: Failed to verify Tibber API access: %v", err)
	}

	return client
}

// VerifyAccess checks if we can access the Tibber API and if the home ID exists
func (c *Client) VerifyAccess() error {
	query := `{
		viewer {
			homes {
				id
				features {
					realTimeConsumptionEnabled
				}
			}
		}
	}`

	// Create request body
	body, err := json.Marshal(map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return fmt.Errorf("error marshaling query: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.tibber.com/v1-beta/gql", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.WebsocketClient.Config.Token))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var result struct {
		Data struct {
			Viewer struct {
				Homes []struct {
					ID       string `json:"id"`
					Features struct {
						RealTimeConsumptionEnabled bool `json:"realTimeConsumptionEnabled"`
					} `json:"features"`
				} `json:"homes"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	// Check for API errors
	if len(result.Errors) > 0 {
		return fmt.Errorf("API error: %s", result.Errors[0].Message)
	}

	// Verify home ID exists and has real-time consumption enabled
	homeFound := false
	for _, home := range result.Data.Viewer.Homes {
		if home.ID == c.WebsocketClient.Config.Id {
			homeFound = true
			if !home.Features.RealTimeConsumptionEnabled {
				return fmt.Errorf("real-time consumption is not enabled for home ID %s", home.ID)
			}
			log.Printf("Successfully verified access to home ID %s with real-time consumption enabled", home.ID)
			break
		}
	}

	if !homeFound {
		return fmt.Errorf("home ID %s not found in account", c.WebsocketClient.Config.Id)
	}

	return nil
}
