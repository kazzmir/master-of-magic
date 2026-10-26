module github.com/kazzmir/master-of-magic/util/combat-simulator/server

go 1.27.0

require (
	github.com/coder/websocket v1.8.15
	github.com/kazzmir/master-of-magic v0.0.0
	github.com/sendgrid/sendgrid-go v3.16.1+incompatible
	golang.org/x/crypto v0.58.0
)

replace github.com/kazzmir/master-of-magic => ../../..

require (
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)
