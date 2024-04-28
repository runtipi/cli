use std::io::{Error, ErrorKind};

use crate::args::{AppCommand, AppSubcommand};
use crate::utils::env::{get_env_value, EnvMap};
use jsonwebtoken::{encode, Algorithm, EncodingKey, Header};
use serde::{Deserialize, Serialize};

use crate::components::spinner;
use crate::utils::constants::DEFAULT_NGINX_PORT;
use reqwest::blocking::{Client, Response};

pub fn run(args: AppCommand, env_map: EnvMap) {
    let base_url = format!(
        "http://{}:{}/worker-api/apps",
        env_map.get("INTERNAL_IP").unwrap_or(&"localhost".to_string()),
        env_map.get("NGINX_PORT").unwrap_or(&DEFAULT_NGINX_PORT.to_string()),
    );

    match args.subcommand {
        AppSubcommand::Start(args) => {
            let spin = spinner::new(&format!("Starting app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "start"));
            let error_message = format!("Failed to start app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App started successfully!");
                    } else {
                        spin.fail(&error_message);
                    }
                }
                Err(err) => {
                    spin.fail("Failed to start app.");
                    println!("{}", format!("Error: {}", err));
                }
            }
            spin.finish();
        }
        AppSubcommand::Stop(args) => {
            let spin = spinner::new(&format!("Stopping app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "stop"));
            let error_message = format!("Failed to stop app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App stopped successfully!");
                    } else {
                        spin.fail(&error_message);
                    }
                }
                Err(err) => {
                    spin.fail("Failed to stop app.");
                    println!("{}", format!("Error: {}", err));
                }
            }
            spin.finish();
        }
        AppSubcommand::Uninstall(args) => {
            let spin = spinner::new(&format!("Uninstalling app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "uninstall"));
            let error_message = format!("Failed to uninstall app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App uninstalled successfully!");
                    } else {
                        spin.fail(&error_message);
                    }
                }
                Err(err) => {
                    spin.fail("Failed to uninstall app.");
                    println!("{}", format!("Error: {}", err));
                }
            }
            spin.finish();
        }
        AppSubcommand::Reset(args) => {
            let spin = spinner::new(&format!("Resetting app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "reset"));
            let error_message = format!("Failed to reset app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App reset successfully!");
                        spin.finish();
                    } else {
                        spin.fail(&error_message);
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to reset app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        }
        AppSubcommand::Update(args) => {
            let spin = spinner::new(&format!("Updating app {}...", args.id));
            let api_response = api_request(format!("{}/{}/{}", base_url, args.id, "update"));
            let error_message = format!("Failed to update app {}. See logs/error.log for more details.", args.id);

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("App updated successfully!");
                        spin.finish();
                    } else {
                        spin.fail(&error_message);
                        spin.finish();
                    }
                }
                Err(err) => {
                    spin.fail("Failed to update app.");
                    spin.finish();
                    println!("{}", format!("Error: {}", err));
                }
            }
        }
        AppSubcommand::StartAll(_) => {
            let spin = spinner::new("Starting all apps...");
            let api_response = api_request(format!("{}/{}", base_url, "start-all"));
            let error_message = format!("Failed to start apps. See logs/error.log for more details.");

            match api_response {
                Ok(response) => {
                    if response.status().is_success() {
                        spin.succeed("All apps started successfully!!");
                        spin.finish();
                    } else {
                        spin.fail(&error_message);
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

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    sub: String,
}

fn create_client() -> Result<Client, Error> {
    let client = Client::builder().build();

    match client {
        Ok(c) => Ok(c),
        Err(err) => return Err(Error::new(ErrorKind::Other, format!("{}", err))),
    }
}

fn create_token() -> Result<String, Error> {
    let claims = Claims { sub: "1".to_string() };

    let jwt_secret = get_env_value("JWT_SECRET");

    match jwt_secret {
        Some(secret) => {
            let encoding_key = EncodingKey::from_secret(secret.as_ref());
            let encoded = encode(&Header::new(Algorithm::HS256), &claims, &encoding_key);

            match encoded {
                Ok(t) => Ok(t),
                Err(err) => Err(Error::new(ErrorKind::Other, format!("Error creating token: {:?}", err))),
            }
        }
        None => Err(Error::new(ErrorKind::Other, "JWT_SECRET not found in environment variables")),
    }
}

fn api_request(url: String) -> Result<Response, Error> {
    let client = create_client()?;

    let token = create_token()?;
    let auth_token = format!("Bearer {}", token);
    let response = client.post(url).header("Authorization", auth_token).send();

    match response {
        Ok(r) => Ok(r),
        Err(err) => Err(Error::new(ErrorKind::Other, format!("{}", err))),
    }
}
