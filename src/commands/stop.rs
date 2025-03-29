use crate::components::spinner;

pub fn run() {
    let spin = spinner::new("");

    let root_folder: PathBuf = current_dir().expect("Unable to get current directory");
    let project_name = root_folder
        .file_name()
        .map(|name| name.to_string_lossy().to_string())
        .unwrap_or_else(|| "runtipi".to_string());

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

    let container_names = vec![
        // Legacy naming
        "tipi-reverse-proxy",
        "tipi-docker-proxy",
        "tipi-db",
        "tipi-redis",
        "tipi-worker",
        "tipi-dashboard",
        // New naming
        "runtipi",
        "runtipi-reverse-proxy",
        "runtipi-db",
        "runtipi-redis",
        "runtipi-queue",
        // New-new naming
        format!("{}-runtipi-1", project_name),
        format!("{}-runtipi-reverse-proxy-1", project_name),
        format!("{}-runtipi-db-1", project_name),
        format!("{}-runtipi-queue-1", project_name),
    ];

    for container_name in container_names {
        let _ = std::process::Command::new("docker").arg("stop").arg(container_name).output();
        let _ = std::process::Command::new("docker").arg("rm").arg(container_name).output();
    }

    spin.succeed("Tipi successfully stopped");
}
