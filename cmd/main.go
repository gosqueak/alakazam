package main

import (
	"crypto/rsa"
	"fmt"
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
	tm, err := team.Download(os.Getenv("TEAMFILE_URL"))
	if err != nil {
		panic(err)
	}

	alakazam := tm.Member("alakazam")
	steelix := tm.Member("steelix")

	db := database.Load(database.DbFileName)
	defer db.Close()

	pKey, err := kit.Retry[*rsa.PublicKey](3, rs256.FetchRsaPublicKey, steelix.Url.String()+"/jwtkeypub")
	if err != nil {
		panic(fmt.Errorf("could not fetch RSA public key: %w", err))
	}

	aud := jwt.NewAudience(pKey, alakazam.JWTInfo.AudienceName)

	apiServ := api.NewServer(alakazam.ListenAddress, db, aud, relay.NewRelay(db, aud))

	apiServ.Run()
}
