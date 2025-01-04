use reqwest::Method;

use crate::args::{AppCommand, AppSubcommand};
use crate::components::spinner::{self, CustomSpinner};
use crate::utils::api::api_request;
use crate::utils::env::EnvMap;

use crate::utils::constants::DEFAULT_NGINX_PORT;

fn handle_api_response(
    spin: CustomSpinner,
    api_response: Result<reqwest::blocking::Response, std::io::Error>,
    success_message: &str,
    error_message: &str,
) {
    match api_response {
        Ok(response) => {
            if response.status().is_success() {
                spin.succeed(success_message);
            } else {
                println!("Error code: {}", response.status());
                spin.fail(error_message);
            }
        }
        Err(err) => {
            spin.fail(error_message);
            println!("Error: {}", err);
        }
    }
    spin.finish();
}

pub fn run(args: AppCommand, env_map: EnvMap) {
    let base_url = format!(
        "http://{}:{}/api/app-lifecycle",
        env_map.get("INTERNAL_IP").unwrap_or(&"localhost".to_string()),
        env_map.get("NGINX_PORT").unwrap_or(&DEFAULT_NGINX_PORT.to_string()),
    );

    match args.subcommand {
        AppSubcommand::Start(args) => {
            let spin = spinner::new(&format!("Starting app {}...", args.urn));
            let url = format!("{}/{}/{}", base_url, args.urn, "start");
            println!("url: {}", url);
            let api_response = api_request(url, Method::POST, "{}");
            let error_message = format!("Failed to start app {}. See logs/error.log for more details.", args.urn);
            handle_api_response(spin, api_response, "App started successfully!", &error_message);
        }
        AppSubcommand::Stop(args) => {
            let spin = spinner::new(&format!("Stopping app {}...", args.urn));
            let url = format!("{}/{}/{}", base_url, args.urn, "stop");
            let api_response = api_request(url, Method::POST, "{}");
            let error_message = format!("Failed to stop app {}. See logs/error.log for more details.", args.urn);
            handle_api_response(spin, api_response, "App stopped successfully!", &error_message);
        }
        AppSubcommand::Uninstall(args) => {
            let spin = spinner::new(&format!("Uninstalling app {}...", args.urn));
            let url = format!("{}/{}/{}", base_url, args.urn, "uninstall");
            let api_response = api_request(url, Method::DELETE, "{\"removeBackups\": false}");
            let error_message = format!("Failed to uninstall app {}. See logs/error.log for more details.", args.urn);
            handle_api_response(spin, api_response, "App uninstalled successfully!", &error_message);
        }
        AppSubcommand::Reset(args) => {
            let spin = spinner::new(&format!("Resetting app {}...", args.urn));
            let url = format!("{}/{}/{}", base_url, args.urn, "reset");
            let api_response = api_request(url, Method::POST, "{}");
            let error_message = format!("Failed to reset app {}. See logs/error.log for more details.", args.urn);
            handle_api_response(spin, api_response, "App reset successfully!", &error_message);
        }
        AppSubcommand::Update(args) => {
            let spin = spinner::new(&format!("Updating app {}...", args.urn));
            let url = format!("{}/{}/{}", base_url, args.urn, "update");
            let api_response = api_request(url, Method::PATCH, "{\"performBackup\": true}");
            let error_message = format!("Failed to update app {}. See logs/error.log for more details.", args.urn);
            handle_api_response(spin, api_response, "App updated successfully!", &error_message);
        }
        AppSubcommand::StartAll(_) => {
            panic!("Not implemented yet");
        }
    }
}
