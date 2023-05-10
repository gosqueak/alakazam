package main

import (
	"crypto/rsa"

	"github.com/gosqueak/alakazam/api"
	"github.com/gosqueak/alakazam/database"
	"github.com/gosqueak/alakazam/relay"
	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
	"github.com/gosqueak/jwt/rs256"
)

const (
	Addr            = "0.0.0.0:8082"
	AuthServerUrl   = "http://0.0.0.0:8081"
	JwtKeyPublicUrl = AuthServerUrl + "/jwtkeypub"
	JwtActorName    = "MSGSERVICE"
)

func main() {
	db := database.Load(database.DbFileName)
	defer db.Close()

	pKey, err := kit.Retry[*rsa.PublicKey](3, rs256.FetchRsaPublicKey, []any{JwtKeyPublicUrl})
	if err != nil {
		panic("Could not fetch RSA public key")
	}

	aud := jwt.NewAudience(pKey, JwtActorName)

	apiServ := api.NewServer(Addr, db, aud, relay.NewRelay(db, aud))

	apiServ.Run()
}
