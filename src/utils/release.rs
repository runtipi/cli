use std::{
    env::current_dir,
    fs::File,
    io::{Error, ErrorKind},
};

use self_update::{self_replace::self_replace, update::Release};
use serde::Deserialize;

use super::{api::create_client, system::get_architecture};

pub fn is_major_bump(current_version: &str, new_version: &str) -> bool {
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

#[derive(Deserialize, Debug)]
struct GithubRelease {
    tag_name: String,
}

pub fn get_latest_release() -> Result<String, Error> {
    let url = "https://api.github.com/repos/runtipi/runtipi/releases/latest";

    let http_client = create_client()?;
    let response = http_client.get(url).send();

    match response {
        Ok(response) => {
            if response.status().is_success() {
                let latest = response.json::<GithubRelease>();

                return match latest {
                    Ok(latest) => Ok(latest.tag_name[1..].to_string()),
                    Err(e) => {
                        return Err(Error::new(ErrorKind::Other, format!("Failed to parse latest release: {:?}", e)));
                    }
                };
            } else {
                return Err(Error::new(
                    ErrorKind::Other,
                    format!("Failed to fetch latest release. Status code: {}", response.status()),
                ));
            }
        }
        Err(e) => Err(Error::new(ErrorKind::Other, format!("Error sending request: {:?}", e))),
    }
}

pub fn get_all_releases() -> Result<Vec<Release>, Error> {
    let releases = self_update::backends::github::ReleaseList::configure()
        .repo_owner("runtipi")
        .repo_name("cli")
        .build()
        .map_err(|e| Error::new(ErrorKind::Other, format!("Failed to find releases from GitHub: {:?}", e)))?;

    let fetch_result = releases
        .fetch()
        .map_err(|e| Error::new(ErrorKind::Other, format!("Failed to fetch releases: {:?}", e)))?;

    Ok(fetch_result)
}

pub fn download_release_and_self_relace(release: &Release) -> Result<(), Error> {
    let arch = get_architecture().unwrap_or("x86_64".to_string()).to_string();
    let arch = if arch == "arm64" { "aarch64".to_string() } else { "x86_64".to_string() };

    let asset = release.asset_for(arch.as_str(), Some("linux"));
    let asset = match asset {
        Some(asset) => asset,
        None => {
            return Err(Error::new(
                ErrorKind::Other,
                format!("No asset found for {} {} on release {}", arch, "linux", release.version),
            ));
        }
    };

    let current_dir = current_dir().expect("Unable to get current directory");

    let tmp_dir = tempfile::Builder::new().prefix("self_update").tempdir_in(&current_dir)?;
    let tmp_tarball_path = tmp_dir.path().join(&asset.name);
    let tmp_tarball = File::create(&tmp_tarball_path)?;

    self_update::Download::from_url(&asset.download_url)
        .set_header(reqwest::header::ACCEPT, "application/octet-stream".parse().unwrap())
        .download_to(&tmp_tarball)
        .map_err(|e| Error::new(ErrorKind::Other, format!("Failed to download release: {:?}", e)))?;

    std::process::Command::new("tar")
        .arg("-xzf")
        .arg(&tmp_tarball_path)
        .arg("-C")
        .arg(&current_dir)
        .output()?;

    // asset.name with no extension
    let bin_name = asset.name.split(".").collect::<Vec<&str>>()[0];
    let new_executable_path = current_dir.join(&bin_name);

    std::process::Command::new("chmod").arg("+x").arg(&new_executable_path);

    let result = self_replace(&new_executable_path);
    std::fs::remove_file(new_executable_path)?;

    result
}
