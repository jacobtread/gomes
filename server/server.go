package server

import _ "embed"

//go:embed cert/cert.pem
var CertFile []byte

//go:embed cert/key.pem
var KeyFile []byte

const RedirectorPort = 42127
const GamePort = 14219
