use crate::components::spinner;

pub fn run() {
    let spin = spinner::new("");

    spin.set_message("Stopping containers...");

    let args = vec!["down", "--remove-orphans", "--rmi", "local"];

    let output = std::process::Command::new("docker")
        .arg("compose")
        .args(&args)
        .output()
        .map_err(|e| e.to_string());

    match output {
        Ok(output) => {
            if !output.status.success() {
                spin.fail("Failed to stop containers. Please try to stop them manually");
                spin.finish();

                let stderr = String::from_utf8_lossy(&output.stderr);
                println!("\nDebug: {}", stderr);
                return;
            }
        }
        Err(e) => {
            spin.fail("Failed to stop containers. Please try to stop them manually");
            spin.finish();

            println!("\nDebug: {}", e);
            return;
        }
    }

    spin.succeed("Tipi successfully stopped");
}
