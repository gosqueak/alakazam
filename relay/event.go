package relay

import "encoding/json"

// enum for event body types
const (
	TypeECDHNotification = "en"
)

// Schema

type socketEvent struct {
	TypeName   string `json:"t"`
	ToUserId   string `json:"tu"`
	FromUserId string `json:"fu"`
	Body       string `json:"b"`
}

type eventBody interface {
	ecdhNotification
}

type ecdhNotification struct {
	ExchangeUUID string `json:"e"`
}

func parseBody[T eventBody](s string) T {
	var v T
	err := json.Unmarshal([]byte(s), &v)

	// TODO handle errors
	if err != nil {

	}

	return v
}
