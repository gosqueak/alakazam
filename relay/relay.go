package relay

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gosqueak/alakazam/database"
	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
)

type connectedUser struct {
	id   string
	conn *websocket.Conn
}

type OutboundEnvelope struct {
	toUserId string
	event    SocketEvent
}

type InboundEnvelope struct {
	fromUserId string
	event      SocketEvent
}

func (u connectedUser) closeConnection(closeCode int, closeText string) {
	u.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(closeCode, closeText),
		time.Now().Add(time.Second*3))

	u.conn.Close()
}

type Relay struct {
	db          *sql.DB
	upgrader    websocket.Upgrader
	connected   map[string]connectedUser
	mu          sync.Mutex
	in          chan InboundEnvelope
	out         chan OutboundEnvelope
	jwtAudience jwt.Audience
}

func NewRelay(db *sql.DB, aud jwt.Audience) *Relay {
	r := &Relay{
		db: db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connected:   make(map[string]connectedUser),
		in:          make(chan InboundEnvelope, 5),
		out:         make(chan OutboundEnvelope, 5),
		jwtAudience: aud,
	}

	go r.sendEnvelopes()
	go r.routeEnvelopes()

	return r
}

func (r *Relay) UpgradeHandler(w http.ResponseWriter, req *http.Request) {
	token, err := kit.GetTokenFromCookie(req, r.jwtAudience.Name)

	if err != nil || !r.jwtAudience.IsValid(token) {
		kit.ErrStatusUnauthorized(w)
		return
	}

	conn, err := r.upgrader.Upgrade(w, req, nil)

	if err != nil {
		kit.ErrInternal(w)
		return
	}

	r.connectUser(conn, token.Body.Subject)
}

func (r *Relay) connectUser(conn *websocket.Conn, userId string) {
	user := connectedUser{
		id:   userId,
		conn: conn,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.connected[user.id] = user

	go r.sendDatabaseEvents(user.id)
	go r.readUserEvents(user)
}

func (r *Relay) getUser(userId string) (u connectedUser, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok = r.connected[userId]
	return u, ok
}

func (r *Relay) purgeUser(u connectedUser) {
	u.closeConnection(websocket.CloseNormalClosure, "")

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.connected, u.id)
}

func (r *Relay) sendDatabaseEvents(userId string) {
	storedEnvelopes, err := database.GetStoredEnvelopes(r.db, userId)

	// TODO handle errors
	if err != nil {

	}

	for _, storedEnvelope := range storedEnvelopes {
		e := SocketEvent{}
		// TODO handle errors here
		json.Unmarshal([]byte(storedEnvelope.Event), &e)

		r.out <- OutboundEnvelope{userId, e}
	}
}

func (r *Relay) readUserEvents(u connectedUser) {
	for {
		evt := SocketEvent{}
		err := u.conn.ReadJSON(&evt)

		// TODO error handling
		if err != nil {

		}

		r.in <- InboundEnvelope{u.id, evt}
	}
}

func (r *Relay) sendEnvelopes() {
	for env := range r.out {
		user, ok := r.getUser(env.toUserId)

		if !ok { //user not connected
			go r.storeEnvelope(env)
			continue
		}

		// TODO add error handling below
		go user.conn.WriteJSON(env.event)
	}
}

func (r *Relay) storeEnvelope(env OutboundEnvelope) {
	eventBytes, err := json.Marshal(env.event)

	if err != nil {
		panic(err)
	}

	err = database.StoreEnvelope(r.db, env.toUserId, string(eventBytes))

	//TODO handle errors
	if err != nil {

	}
}

func (r *Relay) routeEnvelopes() {
	for env := range r.in {
		switch env.event.TypeName {
		case TypeEncryptedMessage:
			em := ParseBody[encryptedMessage](env.event.Body)
			r.out <- OutboundEnvelope{em.ToUserId, env.event}
		case TypeConversationRequest:
			cr := ParseBody[conversationRequest](env.event.Body)
			r.out <- OutboundEnvelope{cr.ToUserId, env.event}
		}
	}
}
