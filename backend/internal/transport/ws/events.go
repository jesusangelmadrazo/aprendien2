package ws

import "encoding/json"

// Event is the envelope for every server->client and client->server message.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ClientMessage is used to decode incoming messages before dispatching by type.
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type LocationUpdatePayload struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type LessonOfferResponsePayload struct {
	OfferID  string `json:"offer_id"`
	Accepted bool   `json:"accepted"`
}
