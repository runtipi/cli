use reqwest::Method;

use crate::args::{AppCommand, AppSubcommand};
use crate::utils::api::api_request;
use crate::utils::env::EnvMap;

use crate::components::spinner;
use crate::utils::constants::DEFAULT_NGINX_PORT;

pub fn run(args: AppCommand, env_map: EnvMap) {
    let base_url = format!(
        "http://{}:{}/api/app-lifecycle",
        env_map.get("INTERNAL_IP").unwrap_or(&"localhost".to_string()),
        env_map.get("NGINX_PORT").unwrap_or(&DEFAULT_NGINX_PORT.to_string()),
    );

    match args.subcommand {
        AppSubcommand::Start(args) => {
            let spin = spinner::new(&format!("Starting app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "start"), Method::POST, "{}");
            let error_message = format!("Failed to start app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App started successfully!");
                    } else {
                        println!("Error code: {}", response.status());
                        spin.fail(&error_message);
                    }
                }
                Err(err) => {
                    spin.fail("Failed to start app.");
                    println!("Error: {}", err);
                }
            }
            spin.finish();
        }
        AppSubcommand::Stop(args) => {
            let spin = spinner::new(&format!("Stopping app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "stop"), Method::POST, "{}");
            let error_message = format!("Failed to stop app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App stopped successfully!");
                    } else {
                        println!("Error code: {}", response.status());
                        spin.fail(&error_message);
                    }
                }
                Err(err) => {
                    spin.fail("Failed to stop app.");
                    println!("Error: {}", err);
                }
            }
            spin.finish();
        }
        AppSubcommand::Uninstall(args) => {
            let spin = spinner::new(&format!("Uninstalling app {}...", args.id));
            let url = format!("{}/{}/{}", base_url, args.id, "uninstall");
            let api_response = api_request(url, Method::DELETE, "{\"removeBackups\": false}");
            let error_message = format!("Failed to uninstall app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App uninstalled successfully!");
                    } else {
                        println!("Error code: {}, {}", response.status(), response.text().unwrap());
                        spin.fail(&error_message);
                    }
                }
                Err(err) => {
                    spin.fail("Failed to uninstall app.");
                    println!("Error: {}", err);
                }
            }
            spin.finish();
        }
        AppSubcommand::Reset(args) => {
            let spin = spinner::new(&format!("Resetting app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "reset"), Method::POST, "{}");
            let error_message = format!("Failed to reset app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App reset successfully!");
                        spin.finish();
                    } else {
                        println!("Error code: {}", response.status());
                        spin.fail(&error_message);
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to reset app.");
                    spin.finish();
                    println!("Error: {}", err);
                }
            }
        }
        AppSubcommand::Update(args) => {
            let spin = spinner::new(&format!("Updating app {}...", args.id));
            let url = format!("{}/{}/{}", base_url, args.id, "update");
            let api_response = api_request(url, Method::PATCH, "{\"performBackup\": true}");
            let error_message = format!("Failed to update app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App updated successfully!");
                        spin.finish();
                    } else {
                        println!("Error code: {}", response.status());
                        spin.fail(&error_message);
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to update app.");
                    spin.finish();
                    println!("Error: {}", err);
                }
            }
        }
        AppSubcommand::StartAll(_) => {
            panic!("Not implemented yet");

            // let spin = spinner::new("Starting all apps...");
            // let api_response = api_request(format!("{}/{}", base_url, "start-all"));
            // let error_message = "Failed to start apps. See logs/error.log for more details.";
            //
            // match api_response {
            //     Ok(response) => {
            //         if response.status().is_success() {
            //             spin.succeed("All apps started successfully!!");
            //             spin.finish();
            //         } else {
            //             spin.fail(error_message);
            //             spin.finish();
            //         }
            //     }
            //     Err(err) => {
            //         spin.fail("Failed to start apps.");
            //         spin.finish();
            //         println!("Error: {}", err);
            //     }
            // }
        }
    }
}
