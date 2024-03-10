use std::collections::HashMap;
use std::env;
use std::path::PathBuf;

use std::io::Error;

use crate::utils::constants::{DEFAULT_NGINX_PORT, DEFAULT_NGINX_PORT_SSL};
use crate::utils::schemas;
use crate::utils::seed::generate_seed;
use crate::utils::system::{derive_entropy, get_architecture, get_internal_ip, get_seed};

use super::constants::{DEFAULT_DOMAIN, DEFAULT_LOCAL_DOMAIN, DEFAULT_POSTGRES_PORT};
use super::schemas::StringOrInt;

pub type EnvMap = HashMap<String, String>;

pub fn get_env_map() -> EnvMap {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");
    let env_file_path = root_folder.join(".env");

    let env_file = std::fs::read_to_string(&env_file_path).expect("Unable to read .env file");
    env_string_to_map(&env_file)
}

pub fn get_env_value(key: &str) -> Option<String> {
    let env_map = get_env_map();
    env_map.get(key).map(|value| value.to_string())
}

pub fn env_string_to_map(env_string: &str) -> EnvMap {
    let mut env_map = std::collections::HashMap::new();

    for line in env_string.lines() {
        // If line is empty or starts with #, skip it
        if line.is_empty() || line.starts_with('#') {
            continue;
        }

        let mut split = line.splitn(2, '=');
        match (split.next(), split.next()) {
            (Some(key), Some(value)) => {
                env_map.insert(key.to_string(), value.to_string());
            }
            _ => {
                eprintln!("Warning: Line '{}' is not in the correct format", line);
            }
        }
    }

    env_map
}

pub fn env_map_to_string(env_map: &EnvMap) -> String {
    let mut env_string = String::new();

    for (key, value) in env_map {
        env_string.push_str(&format!("{}={}\n", key, value));
    }

    env_string
}

pub fn generate_env_file(custom_env_file_path: Option<PathBuf>) -> Result<(), Error> {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");
    let env_file_path = root_folder.join(".env");
    let state_path = root_folder.join("state");
    let settings_file_path = state_path.join("settings.json");

    // Create state folder if it doesn't exist
    std::fs::create_dir_all(state_path)?;

    // Write empty .env file if it doesn't exist
    if !PathBuf::from(&env_file_path).exists() {
        std::fs::write(&env_file_path, "")?;
    }

    // Write empty settings.json file if it doesn't exist
    if !PathBuf::from(&settings_file_path).exists() {
        std::fs::write(&settings_file_path, "{}")?;
    }

    generate_seed(&root_folder)?;

    let env_file = std::fs::read_to_string(&env_file_path)?;
    let env_map = env_string_to_map(&env_file);

    let json_string = std::fs::read_to_string(&settings_file_path)?;
    let parsed_json: schemas::SettingsSchema = serde_json::from_str(&json_string)?;

    let version = std::fs::read_to_string(root_folder.join("VERSION"))?;

    // Create a new env map with the default values
    let mut new_env_map: EnvMap = HashMap::new();

    let seed = get_seed(&root_folder);
    let postgres_password: String = env_map
        .get("POSTGRES_PASSWORD")
        .unwrap_or(&derive_entropy("postgres_password", &seed))
        .to_string();
    let redis_password: String = env_map
        .get("REDIS_PASSWORD")
        .unwrap_or(&derive_entropy("redis_password", &seed))
        .to_string();

    if parsed_json.storage_path.is_some() {
        // Test if the storage path is valid
        let storage_path = PathBuf::from(&parsed_json.storage_path.as_ref().unwrap());

        if !storage_path.exists() {
            return Err(Error::new(
                std::io::ErrorKind::NotFound,
                format!(
                    "Storage path '{}' does not exist on your system. Make sure it is an absolute path or remove it from settings.json.",
                    storage_path.display()
                ),
            ));
        }
    }

    // Insert the default values into the new env map
    new_env_map.insert("INTERNAL_IP".to_string(), parsed_json.internal_ip.unwrap_or(get_internal_ip()));
    new_env_map.insert("ARCHITECTURE".to_string(), get_architecture().unwrap().to_string());
    new_env_map.insert("TIPI_VERSION".to_string(), version);
    new_env_map.insert("ROOT_FOLDER_HOST".to_string(), root_folder.display().to_string());
    new_env_map.insert(
        "NGINX_PORT".to_string(),
        parsed_json.nginx_port.unwrap_or(StringOrInt::from(DEFAULT_NGINX_PORT)).as_string(),
    );
    new_env_map.insert(
        "NGINX_PORT_SSL".to_string(),
        parsed_json
            .nginx_ssl_port
            .unwrap_or(StringOrInt::from(DEFAULT_NGINX_PORT_SSL))
            .as_string(),
    );
    new_env_map.insert(
        "STORAGE_PATH".to_string(),
        parsed_json.storage_path.unwrap_or(root_folder.display().to_string()),
    );
    new_env_map.insert("POSTGRES_PASSWORD".to_string(), postgres_password);
    new_env_map.insert(
        "POSTGRES_PORT".to_string(),
        parsed_json.postgres_port.unwrap_or(StringOrInt::from(DEFAULT_POSTGRES_PORT)).as_string(),
    );
    new_env_map.insert("POSTGRES_HOST".to_string(), "runtipi-db".to_string());
    new_env_map.insert("REDIS_HOST".to_string(), "runtipi-redis".to_string());
    new_env_map.insert("REDIS_PASSWORD".to_string(), redis_password);
    new_env_map.insert("DOMAIN".to_string(), parsed_json.domain.unwrap_or(DEFAULT_DOMAIN.to_string()));
    new_env_map.insert(
        "LOCAL_DOMAIN".to_string(),
        parsed_json.local_domain.unwrap_or(DEFAULT_LOCAL_DOMAIN.to_string()),
    );

    if let Some(custom_env_file_path) = custom_env_file_path {
        let custom_env_file = std::fs::read_to_string(&custom_env_file_path)?;

        let custom_env_map = env_string_to_map(&custom_env_file);

        let mut merged_env_map = new_env_map.clone();

        for (key, value) in custom_env_map {
            merged_env_map.insert(key.clone(), value);
        }

        let merged_env_string = env_map_to_string(&merged_env_map);

        std::fs::write(&env_file_path, merged_env_string)?;
    } else {
        let new_env_string = env_map_to_string(&new_env_map);

        std::fs::write(&env_file_path, new_env_string)?;
    }

    Ok(())
}
