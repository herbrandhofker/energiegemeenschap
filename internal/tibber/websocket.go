package tibber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	log.Printf("Connecting to Tibber WebSocket at %s", url)

	// Create custom dialer with headers
	dialer := websocket.Dialer{
		EnableCompression: true,
		Subprotocols:      []string{"graphql-transport-ws"},
	}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		log.Printf("Failed to connect to Tibber WebSocket: %v", err)
		return
	}
	defer conn.Close()
	log.Printf("Successfully connected to Tibber WebSocket")

	// Send connection init message with empty payload
	initMsg := Message{
		Type:    "connection_init",
		Payload: json.RawMessage(`{}`),
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		log.Printf("Failed to send connection init message: %v", err)
		return
	}
	log.Printf("Sent connection init message")

	// Wait for connection ack
	var ackMsg Message
	if err := conn.ReadJSON(&ackMsg); err != nil {
		log.Printf("Failed to receive connection ack: %v", err)
		return
	}

	if ackMsg.Type != "connection_ack" {
		log.Printf("Received unexpected message type: %s", ackMsg.Type)
		return
	}
	log.Printf("Received connection ack")

	// Prepare subscription payload
	query := fmt.Sprintf(subscriptionQuery, c.WebsocketClient.Config.Id)
	payload := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal subscription payload: %v", err)
		return
	}

	// Send subscription message
	subMsg := Message{
		Type:    "subscribe",
		Id:      "1",
		Payload: payloadBytes,
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		log.Printf("Failed to send subscription message: %v", err)
		return
	}
	log.Printf("Sent subscription message for home ID %s", c.WebsocketClient.Config.Id)

	// Handle incoming messages
	for {
		select {
		case <-ctx.Done():
			// Send close message
			closeMsg := Message{Type: "connection_terminate"}
			if err := conn.WriteJSON(closeMsg); err != nil {
				log.Printf("Failed to send close message: %v", err)
			}
			log.Printf("WebSocket connection closed")
			return
		default:
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("Failed to read message: %v", err)
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
					log.Printf("Failed to unmarshal measurement data: %v", err)
					continue
				}

				if data.Data.LiveMeasurement.Timestamp.IsZero() {
					continue
				}

				c.WebsocketClient.Data <- data.Data.LiveMeasurement
			case "error":
				log.Printf("Received error message: %s", string(msg.Payload))
				continue
			case "complete":
				log.Printf("Received complete message")
				return
			default:
				log.Printf("Received unknown message type: %s", msg.Type)
			}
		}
	}
}
