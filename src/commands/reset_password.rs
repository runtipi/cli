use colored::Colorize;
use std::env;
use std::{fs::write, path::PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

use crate::utils::env::get_env_value;

pub fn run() {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");
    let reset_password_request = write(root_folder.join("state").join("password-change-request"), (SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() + (15 * 60)).to_string());

    match reset_password_request {
        Ok(_) => {
            let internal_ip = get_env_value("INTERNAL_IP").unwrap_or("localhost".to_string());
            let nginx_port = get_env_value("NGINX_PORT").unwrap_or("80".to_string());

            let ip_and_port = format!(
                "Head back to http://{}:{}/reset-password to set your new password.",
                internal_ip, nginx_port
            );

            println!("{} Password reset request created. {}", "✓".green(), ip_and_port)
        }
        Err(e) => {
            println!(
                "{} Unable to create password reset request. You can manually create a file with `echo $(($(date +%s) + 900)) >> {}` to initiate a password reset. Error: {}",
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
