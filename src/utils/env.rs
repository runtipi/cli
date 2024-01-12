use crate::components::spinner;

pub fn generate_env_file() {
    let spin = spinner::new("Generating .env file...");

    // Wait 5 seconds to simulate a long-running task
    std::thread::sleep(std::time::Duration::from_secs(3));

    spin.succeed("Generated .env file");

    spin.set_message("Copying system files...");

    std::thread::sleep(std::time::Duration::from_secs(3));

    spin.fail("Failed to copy system files");

    spin.finish();
}
