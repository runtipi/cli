use colored::Colorize;
use std::env;
use std::{fs::File, path::PathBuf};

use crate::utils::env::get_env_map;

pub fn run() {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");
    let reset_password_request = File::create(root_folder.join("state").join("password-change-request"));

    match reset_password_request {
        Ok(_) => {
            let env_map = get_env_map();

            let ip_and_port = format!(
                "Head back to http://{}:{}/reset-password to set your new password.",
                env_map.get("INTERNAL_IP").unwrap_or(&"localhost".to_string()),
                env_map.get("NGINX_PORT").unwrap_or(&"80".to_string())
            );

            println!("{} Password reset request created. {}", "✓".green(), ip_and_port)
        }
        Err(e) => {
            println!(
                "{} Unable to create password reset request. You can manually create an empty file at {} to initiate a password reset. Error: {}",
                "✗".red(),
                root_folder
                    .join("state")
                    .join("password-change-request")
                    .to_str()
                    .unwrap_or("./state/password-change-request"),
                e
            );
            return;
        }
    }
}
