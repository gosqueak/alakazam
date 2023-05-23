package relay

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/gosqueak/alakazam/database"
	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
)

type connectedUser struct {
	id   string
	conn *websocket.Conn
}

type connectionMap struct {
	users map[string]connectedUser
	mu    sync.Mutex
}

func (c *connectionMap) add(u connectedUser) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[u.id] = u
}

func (c *connectionMap) get(uid string) (u connectedUser, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, ok = c.users[uid]
	return u, ok
}

func (c *connectionMap) purge(u connectedUser) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u.conn.Close()
	delete(c.users, u.id)
}

type Relay struct {
	db             *sql.DB
	upgrader       websocket.Upgrader
	connectedUsers connectionMap
	in             chan socketEvent
	out            chan socketEvent
	jwtAudience    jwt.Audience
}

func NewRelay(db *sql.DB, aud jwt.Audience) *Relay {
	r := &Relay{
		db: db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		connectedUsers: connectionMap{
			users: make(map[string]connectedUser),
		},
		in:          make(chan socketEvent, 5),
		out:         make(chan socketEvent, 5),
		jwtAudience: aud,
	}

	go r.sendSocketEvents()
	go r.receiveSocketEvents()

	return r
}

func (r *Relay) HandleUpgradeConnection(w http.ResponseWriter, req *http.Request) {
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

	defer r.connectedUsers.purge(user)
	r.connectedUsers.add(user)

	// send user events that were stored in db
	go func() {
		eventJSON, _ := database.GetStoredSocketEvents(r.db, userId)
		for _, str := range eventJSON {
			var e socketEvent
			// TODO handle errors here
			json.Unmarshal([]byte(str), &e)
			r.out <- e
		}
	}()

	// block here and continually read events
	for {
		var e socketEvent
		err := user.conn.ReadJSON(&e)

		// TODO error handling
		if err != nil {
			break
		}

		r.in <- e
	}
}

func (r *Relay) sendSocketEvents() {
	for e := range r.out {
		log.Printf("Sending event : %v... => %v... : %v", e.FromUserId[:7], e.ToUserId[:7], e.Body)

		user, ok := r.connectedUsers.get(e.ToUserId)

		if !ok { //user not connected
			go r.storeSocketEvent(e)
			continue
		}

		// TODO add error handling below
		go user.conn.WriteJSON(e)
	}
}

func (r *Relay) storeSocketEvent(e socketEvent) error {
	eventBytes, err := json.Marshal(e)

	if err != nil {
		panic(err)
	}

	return database.StoreSocketEvent(r.db, e.ToUserId, string(eventBytes))

}

func (r *Relay) receiveSocketEvents() {
	for e := range r.in {
		log.Printf("Incoming event : %v... => %v... : %v", e.FromUserId[:7], e.ToUserId[:7], e.Body)
		r.out <- e
	}
}
