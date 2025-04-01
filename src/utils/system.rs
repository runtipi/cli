use hex::encode;

use sha2::{Digest, Sha256};
use std::io::{Error, ErrorKind, Write};
use std::path::Path;
use std::{env, fs};
use std::{fs::File, path::PathBuf};

use super::constants::{DOCKER_COMPOSE_YML, VERSION};

pub fn get_architecture() -> Result<String, Error> {
    #[cfg(target_arch = "aarch64")]
    {
        Ok("arm64".to_string())
    }
    #[cfg(target_arch = "x86_64")]
    {
        Ok("amd64".to_string())
    }
}

pub fn get_internal_ip() -> String {
    match netdev::get_default_interface() {
        Ok(iface) => iface.ipv4[0].addr().to_string(),
        Err(_) => "0.0.0.0".to_string(),
    }
}

pub fn get_seed(root_folder: &Path) -> Result<String, Error> {
    let seed_file_path = root_folder.join("state").join("seed");
    let seed = std::fs::read_to_string(&seed_file_path);

    match seed {
        Ok(seed) => Ok(seed),
        Err(_) => Err(Error::new(
            ErrorKind::Other,
            "Unable to read the seed file. Please run the start command first.".to_string(),
        )),
    }
}

pub fn derive_entropy(entropy: &str, seed: &String) -> String {
    let mut hasher = Sha256::new();
    hasher.update(seed);
    hasher.update(entropy);
    let result = hasher.finalize();
    encode(result)
}

pub fn ensure_docker() -> Result<(), Error> {
    let output = std::process::Command::new("docker").arg("--version").output().map_err(|e| e.to_string());

    match output {
        Ok(output) => {
            if !output.status.success() {
                return Err(Error::new(
                    ErrorKind::Other,
                    "Docker is not installed or user has not the right permissions. See https://docs.docker.com/engine/install/ for more information"
                        .to_string(),
                ));
            }

            // Ensure v28 or higher. Example output: Docker version 27.4.0, build bde2b89
            let version = String::from_utf8_lossy(&output.stdout);

            let version_parts: Vec<&str> = version.split_whitespace().collect();
            let version_number: Vec<&str> = version_parts[2].split(".").collect();

            if version_number[0].parse::<i32>().unwrap() < 28 {
                return Err(Error::new(
                    ErrorKind::Other,
                    "Docker version 28 or higher is required. See https://docs.docker.com/engine/install/ for more information".to_string(),
                ));
            }
        }
        Err(_) => {
            return Err(Error::new(
                ErrorKind::Other,
                "Docker is not installed or user has not the right permissions. See https://docs.docker.com/engine/install/ for more information"
                    .to_string(),
            ));
        }
    }

    let output = std::process::Command::new("docker").arg("compose").arg("version").output();

    match output {
        Ok(output) => {
            if !output.status.success() {
                return Err(Error::new(
                    ErrorKind::Other,
                    "Docker compose plugin is not installed. See https://docs.docker.com/compose/install/linux/ for more information".to_string(),
                ));
            }
        }
        Err(_) => {
            return Err(Error::new(
                ErrorKind::Other,
                "Docker compose plugin is not installed. See https://docs.docker.com/compose/install/linux/ for more information".to_string(),
            ));
        }
    }

    Ok(())
}

/**
* Copy system files to the root folder
*/
pub fn copy_system_files() -> Result<(), Error> {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");

    let mut docker_compose_file = File::create(root_folder.join("docker-compose.yml"))?;
    docker_compose_file.write_all(DOCKER_COMPOSE_YML.as_bytes())?;

    let mut version_file = File::create(root_folder.join("VERSION"))?;
    version_file.write_all(VERSION.as_bytes())?;

    // Create the base folders
    fs::create_dir_all(root_folder.join("apps"))?;
    fs::create_dir_all(root_folder.join("data"))?;
    fs::create_dir_all(root_folder.join("app-data"))?;
    fs::create_dir_all(root_folder.join("state"))?;
    fs::create_dir_all(root_folder.join("repos"))?;
    fs::create_dir_all(root_folder.join("media"))?;
    fs::create_dir_all(root_folder.join("traefik"))?;
    fs::create_dir_all(root_folder.join("user-config"))?;
    fs::create_dir_all(root_folder.join("logs"))?;
    fs::create_dir_all(root_folder.join("backups"))?;

    Ok(())
}

pub fn ensure_file_permissions() -> Result<(), Error> {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");

    let items = vec![
        ("777", vec!["state", "data", "apps", "logs", "traefik", "repos", "user-config", "state"]),
        ("666", vec!["state/settings.json"]),
        ("664", vec![".env", "docker-compose.yml", "VERSION"]),
        ("600", vec!["traefik/shared/acme.json", "state/seed"]),
    ];

    for (perms, paths) in items {
        for path in paths {
            let full_path = root_folder.join(path);
            if !full_path.exists() {
                continue;
            }

            // Try to fix permissions even if the user is not root
            let chmod_status = std::process::Command::new("chmod").arg("-Rf").arg(perms).arg(&full_path).output()?;

            if !chmod_status.status.success() {
                return Err(Error::new(
                    ErrorKind::Other,
                    format!("{} has incorrect permissions. Please run the CLI as root to fix this.", path),
                ));
            }
        }
    }

    Ok(())
}
