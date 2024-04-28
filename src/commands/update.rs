use reqwest::blocking::Client;
use std::path::PathBuf;
use std::{env::current_dir, fs::File};

use self_update::self_replace::self_replace;
use serde::Deserialize;

use crate::components::console_box::ConsoleBox;
use crate::utils::env::get_env_value;
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

fn is_major_bump(current_version: &str, new_version: &str) -> bool {
    if new_version == "nightly" {
        return false;
    }

    let current_version = current_version.split(".").collect::<Vec<&str>>();
    let new_version = new_version.split(".").collect::<Vec<&str>>();

    if current_version[0] < new_version[0] {
        return true;
    }

    false
}

pub fn run(args: UpdateArgs) {
    let spin = spinner::new("");

    spin.set_message("Grabbing releases from GitHub");

    let releases = self_update::backends::github::ReleaseList::configure()
        .repo_owner("runtipi")
        .repo_name("cli")
        .build();

    let releases = match releases {
        Ok(releases) => releases.fetch(),
        Err(e) => {
            spin.fail("Failed to find releases from GitHub");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    };

    let fetch_result = match releases {
        Ok(releases) => releases,
        Err(e) => {
            spin.fail("Failed to fetch releases");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    };

    // Find args.version in releases
    // If args.version is latest, use the latest non-prerelease version
    let wanted_version = if args.version == "latest" {
        let url = "https://api.github.com/repos/runtipi/runtipi/releases/latest";

        let http_client = Client::builder().user_agent("reqwest").build();
        let http_client = match http_client {
            Ok(client) => client,
            Err(e) => {
                spin.fail("Failed to create HTTP client");
                spin.finish();
                println!("\nError: {}", e);
                return;
            }
        };

        let response = http_client.get(url).send();

        let response = match response {
            Ok(response) => response,
            Err(e) => {
                spin.fail("Failed to fetch latest release");
                spin.finish();
                println!("\nError: {}", e);
                return;
            }
        };

        if response.status().is_success() {
            let latest = match response.json::<GithubRelease>() {
                Ok(latest) => latest,
                Err(e) => {
                    spin.fail("Failed to parse latest release");
                    spin.finish();
                    println!("\nError: {}", e);
                    return;
                }
            };
            latest.tag_name[1..].to_string()
        } else {
            spin.fail(format!("Failed to fetch latest release. Status code: {}", response.status()).as_str());
            spin.finish();

            return;
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

    let release = fetch_result.iter().find(|r| r.version.as_str() == wanted_version);

    let release = match release {
        Some(release) => {
            spin.succeed(format!("Found version {}", release.version).as_str());
            release
        }
        None => {
            spin.fail(format!("Version {} not found", wanted_version).as_str());
            spin.finish();
            return;
        }
    };

    let arch = get_architecture().unwrap_or("x86_64".to_string()).to_string();
    let arch = if arch == "arm64" { "aarch64".to_string() } else { "x86_64".to_string() };

    spin.set_message(format!("Downloading {} release", arch).as_str());

    let asset = release.asset_for(arch.as_str(), Some("linux"));
    let asset = match asset {
        Some(asset) => asset,
        None => {
            spin.fail(format!("No asset found for {} {} on release {}", arch, "linux", release.version).as_str());
            spin.finish();
            return;
        }
    };

    let current_dir = current_dir().expect("Unable to get current directory");

    let tmp_dir = tempfile::Builder::new().prefix("self_update").tempdir_in(&current_dir);
    let tmp_dir = match tmp_dir {
        Ok(tmp_dir) => tmp_dir,
        Err(e) => {
            spin.fail("Failed to create temporary directory");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    };

    let tmp_tarball_path = tmp_dir.path().join(&asset.name);
    let tmp_tarball = File::create(&tmp_tarball_path);
    let tmp_tarball = match tmp_tarball {
        Ok(tmp_tarball) => tmp_tarball,
        Err(e) => {
            spin.fail("Failed to create temporary tarball");
            spin.finish();
            println!("\nError: {}", e);
            return;
        }
    };

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

    match args.env_file {
        Some(env_file) => {
            run_args.push("--env-file".to_string());
            run_args.push(env_file.display().to_string());
        }
        None => {}
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

    let ip_and_port = format!("Visit http://{}:{} to access the dashboard", internal_ip, nginx_port,);

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
