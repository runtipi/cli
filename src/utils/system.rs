use hex::encode;

use sha2::{Digest, Sha256};
use std::path::PathBuf;

use get_if_addrs::get_if_addrs;

pub fn get_architecture() -> Result<String, String> {
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
    if let Ok(ifaces) = get_if_addrs() {
        for iface in ifaces {
            if iface.is_loopback() || iface.addr.ip().is_loopback() {
                continue;
            }
            if let get_if_addrs::IfAddr::V4(ref ifv4) = iface.addr {
                // Skip over loopback and check for IPv4
                if !ifv4.ip.is_loopback() {
                    return ifv4.ip.to_string();
                }
            }
        }
    }

    "0.0.0.0".to_string()
}

pub fn get_seed(root_folder: &String) -> String {
    let seed_file_path = PathBuf::from(&root_folder).join("state").join("seed");
    let seed = std::fs::read_to_string(&seed_file_path).unwrap_or_default();
    seed
}

pub fn derive_entropy(entropy: &str, seed: &String) -> String {
    let mut hasher = Sha256::new();
    hasher.update(seed);
    hasher.update(entropy);
    let result = hasher.finalize();
    encode(result)
}
