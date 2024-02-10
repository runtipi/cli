use crate::components::spinner;
use crate::utils::env;
use std::env::set_current_dir;
use std::path::Path;

pub fn run() {
    let spin = spinner::new("");

    spin.set_message("Updating app store repository...");
    std::thread::sleep(std::time::Duration::from_millis(500));

    spin.set_message("Finding repository from .env");
    let repo_id = env::get_env_value("APPS_REPO_ID").unwrap_or("29ca930bfdaffa1dfabf5726336380ede7066bc53297e3c0c868b27c97282903".to_string());
    let path = format!("repos/{}", repo_id);
    let repo_path = Path::new(&path);
    set_current_dir(&repo_path).unwrap();
    spin.succeed("Repository found");

    spin.set_message("Resetting git repository...");
    let reset_command = std::process::Command::new("git").arg("reset").args(["--hard"]).output().map_err(|e| e.to_string());

    match reset_command {
        Ok(output) => {
            if !output.status.success() {
                spin.fail("Failed to reset repository. Please check the permissions.");
                spin.finish();

                let stderr = String::from_utf8_lossy(&output.stderr);
                println!("\nDebug: {}", stderr);
                return;
            }
        }
        Err(e) => {
            spin.fail("Failed to reset repository. Please check the permissions.");
            spin.finish();

            println!("\nDebug: {}", e);
            return;
        }
    }
    spin.succeed("Repository reset");

    spin.set_message("Pulling updates...");
    let pull_command = std::process::Command::new("git").arg("pull").output().map_err(|e| e.to_string());

    match pull_command {
        Ok(output) => {
            if !output.status.success() {
                spin.fail("Failed to git pull. Please check the permissions.");
                spin.finish();

                let stderr = String::from_utf8_lossy(&output.stderr);
                println!("\nDebug: {}", stderr);
                return;
            }
        }
        Err(e) => {
            spin.fail("Failed to git pull. Please check the permissions.");
            spin.finish();

            println!("\nDebug: {}", e);
            return;
        }
    }
    spin.succeed("Updates pulled");

    spin.succeed("Successfully updated app store repository!")
}