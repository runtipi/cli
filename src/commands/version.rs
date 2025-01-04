use crate::utils::env::EnvMap;

pub fn run(env_map: EnvMap) {
    let version = env_map.get("TIPI_VERSION");

    match version {
        Some(version) => println!("{}", version),
        None => println!("Unknown"),
    }
}
