use crate::args::{AppCommand, AppSubcommand};
use reqwest::{Client, Response, Error};
use crate::components::spinner;

pub async fn run(args: AppCommand) {
    let base_url = "http://localhost/worker/api/apps/";
    let auth_token = "secret";
    match args.subcommand {
        AppSubcommand::Start(id) => {
            let spin = spinner::new(&format!("Starting app {}...", id.id));
            let api_response = api_request("start".to_owned(), base_url.to_owned(), auth_token.to_owned(), id.id, false).await;
            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App started successfully!");
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to start app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        },
        AppSubcommand::Stop(id) => {
            let spin = spinner::new(&format!("Stopping app {}...", id.id));
            let api_response = api_request("stop".to_owned(), base_url.to_owned(), auth_token.to_owned(), id.id, false).await;
            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App stopped successfully!");
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to stop app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        },
        AppSubcommand::Uninstall(id) => {
            let spin = spinner::new(&format!("Uninstalling app {}...", id.id));
            let api_response = api_request("uninstall".to_owned(), base_url.to_owned(), auth_token.to_owned(), id.id, false).await;
            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App uninstalled successfully!");
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to uninstall app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        },
        AppSubcommand::Reset(id) => {
            let spin = spinner::new(&format!("Resetting app {}...", id.id));
            let api_response = api_request("reset".to_owned(), base_url.to_owned(), auth_token.to_owned(), id.id, false).await;
            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App reset successfully!");
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to reset app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        },
        AppSubcommand::Update(id) => {
            let spin = spinner::new(&format!("Updating app {}...", id.id));
            let api_response = api_request("update".to_owned(), base_url.to_owned(), auth_token.to_owned(), id.id, false).await;
            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App updated successfully!");
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to update app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        },
        AppSubcommand::StartAll(_) => {
            let spin = spinner::new("Starting all apps...");
            let api_response = api_request("start-all".to_owned(), base_url.to_owned(), auth_token.to_owned(), "none".to_owned(), true).await;
            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("All apps started successfully!!");
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to start apps.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        }
    }
}

async fn api_request(request_type:String, url:String, auth_token:String, app_id:String, start_all_case:bool) -> Result<Response, Error> {
    let client = Client::new();
    let final_url: String;
    if start_all_case == true {
        final_url = format!("{}{}", url, request_type);
    } else {
        final_url = format!("{}{}/{}", url, app_id, request_type);
    }
    let response = client.post(final_url).header("Authorization", auth_token).send().await?;
    Ok(response)
}