package tibber

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

const subscriptionQuery = `subscription {
	liveMeasurement(homeId: "%s") {
		timestamp
		power
		powerProduction
		minPower
		averagePower
		maxPower
		maxPowerProduction
		accumulatedConsumption
		accumulatedProduction
		lastMeterConsumption
		lastMeterProduction
		currentL1
		currentL2
		currentL3
		voltagePhase1
		voltagePhase2
		voltagePhase3
		signalStrength
	}
}`

func (c *Client) Subscribe(ctx context.Context) {
	defer c.Wg.Done()

	// Create WebSocket connection
	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Bearer %s", c.WebsocketClient.Config.Token))
	header.Add("User-Agent", "TibberClient/1.0 (Go)")

	url := fmt.Sprintf("wss://%s%s", c.WebsocketClient.Config.Host, c.WebsocketClient.Config.Path)

	// Create custom dialer with headers
	dialer := websocket.Dialer{
		EnableCompression: true,
		Subprotocols:      []string{"graphql-transport-ws"},
	}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return
	}
	defer conn.Close()

	// Send connection init message with empty payload
	initMsg := Message{
		Type:    "connection_init",
		Payload: json.RawMessage(`{}`),
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		return
	}

	// Wait for connection ack
	var ackMsg Message
	if err := conn.ReadJSON(&ackMsg); err != nil {
		return
	}

	if ackMsg.Type != "connection_ack" {
		return
	}

	// Prepare subscription payload
	query := fmt.Sprintf(subscriptionQuery, c.WebsocketClient.Config.Id)
	payload := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Send subscription message
	subMsg := Message{
		Type:    "subscribe",
		Id:      "1",
		Payload: payloadBytes,
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		return
	}

	// Handle incoming messages
	for {
		select {
		case <-ctx.Done():
			// Send close message
			closeMsg := Message{Type: "connection_terminate"}
			if err := conn.WriteJSON(closeMsg); err != nil {
				return
			}
			return
		default:
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}

			switch msg.Type {
			case "next":
				var data struct {
					Data struct {
						LiveMeasurement Measurement `json:"liveMeasurement"`
					} `json:"data"`
				}

				if err := json.Unmarshal(msg.Payload, &data); err != nil {
					continue
				}

				if data.Data.LiveMeasurement.Timestamp.IsZero() {
					continue
				}

				c.WebsocketClient.Data <- data.Data.LiveMeasurement
			case "error":
				continue
			case "complete":
				return
			}
		}
	}
}
