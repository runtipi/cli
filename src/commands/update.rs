use reqwest::blocking::Client;
use std::path::PathBuf;
use std::{env::current_dir, fs::File};

use self_update::self_replace::self_replace;
use serde::Deserialize;

use crate::components::console_box::ConsoleBox;
use crate::utils::env;
use crate::{components::spinner, utils::system::get_architecture};

#[derive(Deserialize, Debug)]
struct GithubRelease {
    tag_name: String,
}

#[derive(Debug)]
pub struct UpdateArgs {
    pub version: String,
    pub env_file: Option<PathBuf>,
    pub no_permissions: bool,
}

pub fn run(args: UpdateArgs) {
    let spin = spinner::new("");

    spin.set_message("Grabbing releases from GitHub");

    let releases = self_update::backends::github::ReleaseList::configure()
        .repo_owner("runtipi")
        .repo_name("cli")
        .build()
        .unwrap();

    let fetch = releases.fetch().unwrap();

    // Find args.version in releases
    // If args.version is latest, use the latest non-prerelease version
    let wanted_version = if args.version == "latest" {
        let url = "https://api.github.com/repos/runtipi/runtipi/releases/latest";

        let http_client = Client::builder().user_agent("reqwest").build().unwrap();
        let response = http_client.get(url).send().unwrap();

        if response.status().is_success() {
            let latest: GithubRelease = response.json().unwrap();

            latest.tag_name[1..].to_string()
        } else {
            spin.fail("Failed to fetch latest release");
            spin.finish();

            return;
        }
    } else {
        args.version
    };

    let release = fetch.iter().find(|r| r.version.as_str() == wanted_version);

    match release {
        Some(release) => {
            spin.succeed(format!("Found version {}", release.version).as_str());
        }
        None => {
            spin.fail(format!("Version {} not found", wanted_version).as_str());
            spin.finish();
            return;
        }
    }

    let release = release.unwrap();

    let arch = get_architecture().unwrap_or("x86_64".to_string()).to_string();
    let arch = if arch == "arm64" { "aarch64".to_string() } else { "x86_64".to_string() };

    spin.set_message(format!("Downloading {} release", arch).as_str());

    let asset = release.asset_for(arch.as_str(), Some("linux"));

    if asset.is_none() {
        spin.fail(format!("No asset found for {} {} on release {}", arch, "linux", release.version).as_str());
        spin.finish();
        return;
    }

    let asset = asset.unwrap();

    let current_dir = current_dir().expect("Unable to get current directory");

    let tmp_dir = tempfile::Builder::new().prefix("self_update").tempdir_in(&current_dir).unwrap();
    let tmp_tarball_path = tmp_dir.path().join(&asset.name);
    let tmp_tarball = File::create(&tmp_tarball_path).unwrap();

    let output = self_update::Download::from_url(&asset.download_url)
        .set_header(reqwest::header::ACCEPT, "application/octet-stream".parse().unwrap())
        .download_to(&tmp_tarball);

    match output {
        Ok(_) => {
            spin.succeed(format!("Downloaded {}", &asset.name).as_str());
        }
        Err(e) => {
            spin.fail("Failed to download release");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    }

    spin.set_message("Extracting tarball");

    let output = std::process::Command::new("tar")
        .arg("-xzf")
        .arg(&tmp_tarball_path)
        .arg("-C")
        .arg(&current_dir)
        .output();

    match output {
        Ok(_) => {
            spin.succeed("Extracted tarball");
        }
        Err(e) => {
            spin.fail("Failed to extract tarball");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    }

    // asset.name with no extension
    let bin_name = asset.name.split(".").collect::<Vec<&str>>()[0];
    let new_executable_path = current_dir.join(&bin_name);

    spin.set_message("Replacing old CLI");
    std::process::Command::new("chmod").arg("+x").arg(&new_executable_path);

    let result = self_replace(&new_executable_path);
    let _ = std::fs::remove_file(new_executable_path);

    match result {
        Ok(_) => {
            spin.succeed("Tipi updated successfully. Starting new CLI");
        }
        Err(e) => {
            spin.fail("Failed to replace old CLI");
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

    let env_file = args.env_file;

    if env_file.is_some() {
        run_args.push("--env-file".to_string());
        run_args.push(env_file.unwrap().display().to_string());
    }

    // Run command start on new CLI
    let result = std::process::Command::new("./runtipi-cli").args(run_args).output();

    match result {
        Ok(output) => {
            if output.status.success() {
                return;
            } else {
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

    let env_map = env::get_env_map();

    println!("\n");

    let ip_and_port = format!(
        "Visit http://{}:{} to access the dashboard",
        env_map.get("INTERNAL_IP").unwrap(),
        env_map.get("NGINX_PORT").unwrap()
    );

    let box_title = "Runtipi started successfully".to_string();
    let box_body = format!(
        "{}\n\n{}\n\n{}",
        ip_and_port,
        format!("You are now running version {}", release.version),
        "Tipi is entirely written in TypeScript and we are looking for contributors!"
    );

    let console_box = ConsoleBox::new(box_title, box_body, 80, "green".to_string());
    console_box.print();
}
