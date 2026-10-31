package assets

import _ "embed"

//go:embed docker-compose.yml.tmpl
var DockerComposeTemplate []byte
