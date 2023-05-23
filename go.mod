module github.com/gosqueak/alakazam

go 1.20

require (
	github.com/gorilla/websocket v1.5.0
	github.com/gosqueak/apikit v0.0.0-20230512061655-69436fe1a189
	github.com/gosqueak/jwt v0.0.0-20230510165842-1b5cd2f15c4b
	github.com/gosqueak/leader v0.0.0-20230517164329-a629fcaf0fe9
	github.com/mattn/go-sqlite3 v1.14.16
)

replace github.com/gosqueak/apikit => ../apikit

replace github.com/gosqueak/jwt => ../jwt

replace github.com/gosqueak/leader => ../leader
