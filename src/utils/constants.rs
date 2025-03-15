pub const DEFAULT_NGINX_PORT: &str = "80";
pub const DEFAULT_NGINX_PORT_SSL: &str = "443";
pub const DEFAULT_POSTGRES_PORT: &str = "5432";
pub const DEFAULT_DOMAIN: &str = "example.com";
pub const DEFAULT_LOCAL_DOMAIN: &str = "tipi.local";
pub const DOCKER_COMPOSE_YML: &str = include_str!("../assets/docker-compose.yml");
pub const VERSION: &str = include_str!("../assets/VERSION");
pub const DEFAULT_FORWARD_AUTH_URL: &str = "http://runtipi:3000/api/auth/traefik";
