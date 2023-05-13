package main

import (
	"crypto/rsa"
	"os"

	"github.com/gosqueak/alakazam/api"
	"github.com/gosqueak/alakazam/database"
	"github.com/gosqueak/alakazam/relay"
	kit "github.com/gosqueak/apikit"
	"github.com/gosqueak/jwt"
	"github.com/gosqueak/jwt/rs256"
	"github.com/gosqueak/leader/team"
)

func main() {
	tm := team.Download(os.Getenv("TEAMFILE_JSON_URL"))
	alakazam := tm["alakazam"]
	steelix := tm["steelix"]

	db := database.Load(database.DbFileName)
	defer db.Close()

	pKey, err := kit.Retry[*rsa.PublicKey](3, rs256.FetchRsaPublicKey, steelix.Url+"/jwtkeypub")
	if err != nil {
		panic("Could not fetch RSA public key")
	}

	aud := jwt.NewAudience(pKey, alakazam.JWTInfo.AudienceName)

	apiServ := api.NewServer(alakazam.Url, db, aud, relay.NewRelay(db, aud))

	apiServ.Run()
}
