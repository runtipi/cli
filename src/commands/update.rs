use std::path::PathBuf;

use crate::components::console_box::ConsoleBox;
use crate::components::spinner;
use crate::utils::env::get_env_value;
use crate::utils::release::{download_release_and_self_relace, get_all_releases, get_latest_release, is_major_bump};

#[derive(Debug)]
pub struct UpdateArgs {
    pub version: String,
    pub env_file: Option<PathBuf>,
    pub no_permissions: bool,
}

pub fn run(args: UpdateArgs) {
    let spin = spinner::new("");

    spin.set_message("Grabbing releases from GitHub");

    // Find args.version in releases
    // If args.version is latest, use the latest non-prerelease version
    let wanted_version = if args.version == "latest" {
        let latest = get_latest_release();
        match latest {
            Ok(latest) => latest,
            Err(e) => {
                spin.fail("Failed to fetch latest release");
                spin.finish();
                println!("\nError: {}", e);
                return;
            }
        }
    } else if args.version == "nightly" {
        "nightly".to_string()
    } else {
        args.version
    };

    let current_version = get_env_value("TIPI_VERSION").unwrap().replace("v", "");
    if is_major_bump(&current_version, &wanted_version) {
        spin.fail("You are trying to update to a new major version. Please update manually using the update instructions on the website. https://runtipi.io/docs/reference/breaking-updates");
        spin.finish();
        return;
    }

    let releases = match get_all_releases() {
        Ok(releases) => releases,
        Err(e) => {
            spin.fail("Failed to fetch releases");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    };

    spin.succeed("Releases fetched successfully");
    spin.set_message("Finding release");

    let release = match releases.iter().find(|r| r.version.as_str() == wanted_version) {
        Some(release) => {
            spin.succeed(format!("Found version {}", release.version).as_str());
            release
        }
        None => {
            spin.fail(format!("Version {} not found in release list", wanted_version).as_str());
            spin.finish();
            return;
        }
    };

    spin.set_message("Downloading release assets");

    let download = download_release_and_self_relace(release);
    match download {
        Ok(_) => {
            spin.succeed("Tipi updated successfully. Starting new CLI");
        }
        Err(e) => {
            spin.fail("Failed to download release");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    }

    spin.set_message("Starting Tipi... This may take a while.");

    // Start new CLI
    let mut run_args = vec!["start".to_string()];
    if args.no_permissions {
        run_args.push("--no-permissions".to_string());
    }

    if let Some(env_file) = args.env_file {
        run_args.push("--env-file".to_string());
        run_args.push(env_file.display().to_string());
    }

    // Run command start on new CLI
    let result = std::process::Command::new("./runtipi-cli").args(run_args).output();

    match result {
        Ok(output) => {
            if !output.status.success() {
                spin.fail("Failed to start new CLI");
                println!("\nDebug: {}", String::from_utf8_lossy(&output.stderr));
            }
        }
        Err(e) => {
            spin.fail("Failed to start new CLI");
            println!("\nDebug: {}", e);
            return;
        }
    }

    spin.finish();

    println!("\n");

    let internal_ip = get_env_value("INTERNAL_IP").unwrap_or("localhost".to_string());
    let nginx_port = get_env_value("NGINX_PORT").unwrap_or("80".to_string());

    let box_title = "Runtipi started successfully".to_string();

    let ip_and_port = format!("Visit http://{}:{} to access the dashboard", internal_ip, nginx_port);
    let message = format!("You are now running version {}", release.version);
    let shameless_plug = "Tipi is entirely written in TypeScript and we are looking for contributors!";

    let box_body = format!("{}\n\n{}\n\n{}", ip_and_port, message, shameless_plug);

    let console_box = ConsoleBox::new(box_title, box_body, 80, "green".to_string());
    console_box.print();
}
