package relay

import "encoding/json"

// enum for event body types
const (
	TypeEncryptedMessage = "em"
	TypeConversationRequest = "cr"
)

// Schema

type SocketEvent struct {
	TypeName string `json:"t"`
	Body     string `json:"b"`
}

type eventBody interface {
	encryptedMessage | conversationRequest
}

type encryptedMessage struct {
	ToUserId          string `json:"t"`
	B64Ciphertext     string `json:"b"`
	SenderPreKeyId    string `json:"s"`
	RecipientPrekeyId string `json:"r"`
}

type conversationRequest struct {
	ToUserId       string `json:"t"`
	FromUserId     string `json:"f"`
	ConversationId string `json:"c"`
}

//
func ParseBody[T eventBody](s string) T {
	var v T
	err := json.Unmarshal([]byte(s), &v)

	// TODO handle errors
	if err != nil {

	}

	return v
}
