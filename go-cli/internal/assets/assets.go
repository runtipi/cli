package assets

import _ "embed"

//go:embed docker-compose.yml
var DockerCompose []byte

//go:embed VERSION
var Version []byte
