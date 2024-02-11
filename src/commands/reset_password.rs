use std::{fs::File, path::PathBuf};
use std::env;
use colored::Colorize;

pub fn run() {
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");
    let mut _reset_password_request = File::create(root_folder.join("state").join("password-change-request"));
    print!("{}", "✓ ".green());
    println!("Password reset request created. Head back to the dashboard to set a new password.")
}