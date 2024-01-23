use crate::components::spinner;
use std::{env, fs, process};
use std::path::PathBuf;
use std::io::Error;

pub fn run() {
    let spin = spinner::new("");
    let root_folder: PathBuf = env::current_dir().expect("Unable to get current directory");

    spin.set_message("Creating backup directory...");

    if let Err(e) =  fs::create_dir_all(root_folder.join("backups")) {
        spin.fail(e.to_string().as_str());
        spin.finish();
        process::exit(1);
    }

    spin.succeed("Created backup directory");

    spin.set_message("Stopping Tipi...");

    let args = vec!["down", "--remove-orphans", "--rmi", "local"];

    process::Command::new("docker")
        .arg("compose")
        .args(&args)
        .output()
        .expect("Failed to stop containers. Cannot continue with backup");

    spin.succeed("Tipi successfully stopped");

    spin.set_message("Backing up tipi...");

    if let Err(e) = backup(root_folder) {
        spin.fail(e.to_string().as_str());
        spin.finish();
        return;
    }

    spin.succeed("Tipi Backed Up!");
    spin.finish();
}

pub fn backup(root_folder: PathBuf) -> Result<(), Error> {
    let datetime = chrono::Utc::now();
    let parent_folder = root_folder.parent().unwrap();
    let root_folder_string = root_folder.to_str().unwrap();
    let filename = format!("runtipi-backup-{}.tar.gz", datetime.format("%d-%m-%Y"));
    let tar_path = format!("{}/{}", parent_folder.to_str().unwrap(), filename);
    let backups_folder = format!("{}/backups/{}", root_folder_string, filename); 

    env::set_current_dir(&parent_folder)?;

    process::Command::new("tar").args(["-czvf", &filename, &root_folder_string]).output().expect("Failed to run tar command");
    fs::rename(tar_path, backups_folder).expect("Failed to move the tar archive.");

    Ok(())
}