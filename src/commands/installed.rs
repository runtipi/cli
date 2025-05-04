use reqwest::Method;

use crate::components::spinner::{self};
use crate::utils::api::api_request;
use crate::utils::env::EnvMap;

use crate::utils::constants::DEFAULT_NGINX_PORT;

fn handle_installed_response(
    api_response: Result<reqwest::blocking::Response, std::io::Error>
) -> Result<serde_json::Value, String> {
    match api_response {
        Ok(response) => {
            if response.status().is_success() {
                return response.json::<serde_json::Value>().map_err(|e| e.to_string());
            } else {
                return Err(format!("Error code: {}", response.status()));
            }
        }
        Err(err) => {
            return Err(format!("{}", err));
        }
    }
}

pub fn run(env_map: EnvMap) {
    let base_url = format!(
        "http://{}:{}/api/apps/installed",
        env_map.get("INTERNAL_IP").unwrap_or(&"localhost".to_string()),
        env_map.get("NGINX_PORT").unwrap_or(&DEFAULT_NGINX_PORT.to_string()),
    );

    let spin = spinner::new("Getting list of installed apps...");
    let api_response = api_request(base_url, Method::GET, "");
    match handle_installed_response(api_response) {
        Ok(apps_json) => {
            if let Some(installed_apps) = apps_json["installed"].as_array() {
                if installed_apps.is_empty() {
                    spin.succeed("No apps found.");
                } else {
                    spin.succeed("Retrieved installed apps.");
                    for app in installed_apps {
                        if let Some(app_info) = app.get("info") {
                            if let Some(app_id) = app_info.get("urn").and_then(|urn| urn.as_str()) {
                                println!("{}", app_id);
                            } else {
                                eprintln!("Warning: Missing or invalid 'urn' field in app info.");
                            }
                        } else {
                            eprintln!("Warning: Missing 'info' field in app.");
                        }
                    }
                }
            } else {
                spin.fail("Unexpected response format: 'installed' field is missing or invalid.");
            }
        }
        Err(err) => {
            spin.fail("Failed to retrieve installed apps.");
            println!("{}", err);
        }
    }
    spin.finish();
}