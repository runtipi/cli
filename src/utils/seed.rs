use std::{fs, path::PathBuf};

use rand::{distributions::Alphanumeric, Rng};

pub fn generate_seed(root_folder: &PathBuf) -> Result<(), std::io::Error> {
    let seed_path = root_folder.join("state").join("seed");

    // Check if the seed file exists
    if !seed_path.exists() {
        // Generate random bytes (32 characters)
        let rng = rand::thread_rng();
        let random_bytes: String = rng
            .sample_iter(&Alphanumeric)
            .take(32)
            .map(char::from)
            .collect();

        // Write the random bytes to the file
        fs::write(&seed_path, random_bytes)?;
    }

    Ok(())
}
