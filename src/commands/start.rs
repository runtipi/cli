use std::env::current_dir;
use std::path::PathBuf;

use crate::args::StartArgs;
use crate::components::spinner;
use crate::utils::{env, system};

pub fn run(args: StartArgs) {
    println!("Starting Runtipi...\n");

    let spin = spinner::new("");

    // User permissions
    spin.set_message("Checking user permissions");

    if let Err(e) = system::ensure_docker() {
        spin.fail(e.to_string().as_str());
        spin.finish();
        return;
    }

    if let Err(e) = system::ensure_user_and_group() {
        spin.fail("Unable to create user and group. Please run the CLI as root.");
        spin.finish();
        println!("\nError: {}", e);
        return;
    }

    spin.succeed("User permissions are ok");

    // System files
    spin.set_message("Copying system files...");

    if let Err(e) = system::copy_system_files() {
        spin.fail("Failed to copy system files");
        spin.finish();
        println!("\nError: {}", e);
        return;
    }
    spin.succeed("Copied system files");

    // Env file generation
    spin.set_message("Generating .env file...");

    if let Err(e) = env::generate_env_file(args.env_file) {
        spin.fail("Failed to generate .env file");
        spin.finish();
        println!("\nError: {}", e);
        return;
    }

    spin.succeed("Generated .env file");

    spin.set_message("Ensuring file permissions...");

    if !args.no_permissions {
        if let Err(e) = system::ensure_file_permissions() {
            spin.fail(e.to_string().as_str());
            spin.finish();
            return;
        }
    }

    spin.succeed("File permissions ok");

    spin.set_message("Pulling images...");

    let root_folder: PathBuf = current_dir().expect("Unable to get current directory");

    let env_file_path = format!("{}/.env", root_folder.display());
    let output = std::process::Command::new("docker")
        .arg("compose")
        .arg("--env-file")
        .arg(&env_file_path)
        .arg("pull")
        .output();

    match output {
        Ok(output) => {
            if !output.status.success() {
                spin.fail("Failed to pull images");

                let stderr = String::from_utf8_lossy(&output.stderr);
                println!("\nDebug: {}", stderr);
                return;
            }
        }
        Err(e) => {
            spin.fail("Failed to pull images");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    }

    spin.succeed("Images pulled");

    // Stop and remove containers
    spin.set_message("Stopping existing containers...");
    let container_names = vec!["tipi-reverse-proxy", "tipi-db", "tipi-redis", "tipi-worker", "tipi-dashboard"];

    for container_name in container_names {
        let _ = std::process::Command::new("docker").arg("stop").arg(container_name).output();
        let _ = std::process::Command::new("docker").arg("rm").arg(container_name).output();
    }

    spin.succeed("Existing containers stopped");

    spin.set_message("Starting containers...");
    let user_compose_file = root_folder.join("user-config").join("tipi-compose.yml");

    let mut args = vec!["-f".to_string(), root_folder.join("docker-compose.yml").display().to_string()];

    if user_compose_file.exists() {
        args.push("-f".to_string());
        args.push(user_compose_file.display().to_string());
    }

    args.push("--env-file".to_string());
    args.push(env_file_path);
    args.push("up".to_string());
    args.push("--detach".to_string());
    args.push("--remove-orphans".to_string());
    args.push("--build".to_string());

    let output = std::process::Command::new("docker")
        .arg("compose")
        .args(&args)
        .output()
        .map_err(|e| e.to_string());

    match output {
        Ok(output) => {
            if !output.status.success() {
                spin.fail("Failed to start containers");

                let stderr = String::from_utf8_lossy(&output.stderr);
                println!("\nDebug: {}", stderr);
                return;
            }
        }
        Err(e) => {
            spin.fail("Failed to start containers");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    }

    spin.succeed("Containers started");

    spin.finish();
}
