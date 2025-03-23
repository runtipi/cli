use colored::Colorize;

use prettytable::{format, row, Table};
use serde_json::{to_string_pretty, Value};

use crate::utils::{env::EnvMap, system::get_architecture};

pub fn run(env_map: EnvMap) {
    println!("⚠️ Make sure you have started tipi before running this command\n");
    // Gather system information
    let os = std::env::consts::OS;
    let version = sys_info::os_release().unwrap_or_else(|_| "Unknown".to_string());
    let mem = sys_info::mem_info().map(|mi| mi.total).unwrap_or(0);
    let arch = get_architecture().unwrap_or("Unknown".to_string());
    let current_dir = std::env::current_dir().unwrap_or_default();

    // Create a table and add rows with system information
    println!("--- {} ---", "System information".blue());
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_BOX_CHARS);
    table.add_row(row!["OS", os]);
    table.add_row(row!["OS Version", version]);
    table.add_row(row!["Memory (GB)", format!("{:.2}", mem as f64 / 1024.0 / 1024.0)]);
    table.add_row(row!["Architecture", arch]);

    // Does the file user_config/tipi-config.yml exist?
    let config_file = std::path::Path::new("user-config/tipi-compose.yml");

    // Print the table
    table.printstd();
    // println!("\n--- \x1b[94mTipi configuration\x1b[0m ---");
    println!("\n--- {} ---", "Tipi configuration".blue());
    let mut table = Table::new();

    table.set_format(*format::consts::FORMAT_BOX_CHARS);
    table.add_row(row![
        "Custom tipi docker config",
        if config_file.exists() { "Yes".yellow() } else { "No".bright_white() }
    ]);

    table.printstd();

    println!("\n--- {} ---", "Settings.json".blue());
    let settings_file_path = current_dir.join("state").join("settings.json");

    let json_string = std::fs::read_to_string(&settings_file_path).unwrap_or_default();
    let parsed_json: Value = serde_json::from_str(&json_string).unwrap_or_default();

    // Pretty print the JSON
    let pretty_json = to_string_pretty(&parsed_json).unwrap_or_else(|_| {
        eprintln!("Failed to generate pretty JSON.");
        String::new()
    });

    println!("{}", pretty_json);

    println!("\n--- {} ---", "Environment variables".blue());
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_BOX_CHARS);

    let pg_password = env_map
        .get("POSTGRES_PASSWORD")
        .map(|_| "<redacted>".to_string())
        .unwrap_or("Not set".red().to_string());
    let rabbitmq_password = env_map
        .get("RABBITMQ_PASSWORD")
        .map(|_| "<redacted>".to_string())
        .unwrap_or("Not set".red().to_string());
    let jwt_secret = env_map
        .get("JWT_SECRET")
        .map(|_| "<redacted>".to_string())
        .unwrap_or("Not set".red().to_string());
    let domain = env_map.get("DOMAIN").map(|_| "<redacted>").unwrap_or("Not set");

    table.add_row(row!["POSTGRES_PASSWORD", pg_password]);
    table.add_row(row!["RABBITMQ_PASSWORD", rabbitmq_password]);
    table.add_row(row!["APPS_REPO_ID", env_map.get("APPS_REPO_ID").unwrap_or(&"Not set".red().to_string())]);
    table.add_row(row![
        "APPS_REPO_URL",
        env_map.get("APPS_REPO_URL").unwrap_or(&"Not set".red().to_string())
    ]);
    table.add_row(row!["TIPI_VERSION", env_map.get("TIPI_VERSION").unwrap_or(&"Not set".red().to_string())]);
    table.add_row(row!["INTERNAL_IP", env_map.get("INTERNAL_IP").unwrap_or(&"Not set".red().to_string())]);
    table.add_row(row!["ARCHITECTURE", env_map.get("ARCHITECTURE").unwrap_or(&"Not set".red().to_string())]);
    table.add_row(row!["JWT_SECRET", jwt_secret]);
    table.add_row(row![
        "ROOT_FOLDER_HOST",
        env_map.get("ROOT_FOLDER_HOST").unwrap_or(&"Not set".red().to_string())
    ]);
    table.add_row(row![
        "RUNTIPI_APP_DATA_PATH",
        env_map.get("RUNTIPI_APP_DATA_PATH").unwrap_or(&"Not set".red().to_string())
    ]);
    table.add_row(row!["NGINX_PORT", env_map.get("NGINX_PORT").unwrap_or(&"Not set".red().to_string())]);
    table.add_row(row![
        "NGINX_PORT_SSL",
        env_map.get("NGINX_PORT_SSL").unwrap_or(&"Not set".red().to_string())
    ]);
    table.add_row(row!["DOMAIN", domain]);
    table.add_row(row!["POSTGRES_HOST", env_map.get("POSTGRES_HOST").unwrap_or(&"Not set".to_string())]);
    table.add_row(row!["POSTGRES_DBNAME", env_map.get("POSTGRES_DBNAME").unwrap_or(&"Not set".to_string())]);
    table.add_row(row![
        "POSTGRES_USERNAME",
        env_map.get("POSTGRES_USERNAME").unwrap_or(&"Not set".to_string())
    ]);
    table.add_row(row![
        "POSTGRES_PORT",
        env_map.get("POSTGRES_PORT").unwrap_or(&"Not set".red().to_string())
    ]);
    table.add_row(row!["RABBITMQ_HOST", env_map.get("RABBITMQ_HOST").unwrap_or(&"Not set".to_string())]);
    table.add_row(row![
        "RABBITMQ_USERNAME",
        env_map.get("RABBITMQ_USERNAME").unwrap_or(&"Not set".to_string())
    ]);
    table.add_row(row!["DEMO_MODE", env_map.get("DEMO_MODE").unwrap_or(&"Not set".to_string())]);
    table.add_row(row!["LOCAL_DOMAIN", env_map.get("LOCAL_DOMAIN").unwrap_or(&"Not set".to_string())]);

    table.printstd();

    println!("\n--- {} ---", "Docker containers".blue());
    table = Table::new();
    table.set_format(*format::consts::FORMAT_BOX_CHARS);
    let containers = std::process::Command::new("docker")
        .arg("ps")
        .arg("-a")
        .arg("--filter")
        .arg("name=runtipi")
        .arg("--format")
        .arg("{{.Names}} {{.Status}}")
        .output()
        .map_err(|e| e.to_string());

    match containers {
        Ok(output) => {
            if output.status.success() {
                let containers = String::from_utf8_lossy(&output.stdout);
                let containers = containers
                    .split("\n")
                    .filter(|s| !s.is_empty())
                    .map(|s| s.to_string())
                    .collect::<Vec<String>>();
                for container in containers {
                    let array = container.split(" ").collect::<Vec<&str>>();

                    let status = if array[1].contains("Up") { array[1].green() } else { array[1].red() };

                    table.add_row(row![array[0], status]);
                }
            } else {
                table.add_row(row!["No containers found"]);
            }
        }
        Err(_) => {
            table.add_row(row!["No containers found"]);
        }
    };

    table.printstd();
    println!("^ If a container is not 'Up', you can run the command `docker logs <container_name>` to see the logs of that container.");
}
