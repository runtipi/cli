use crate::args::StartArgs;
use crate::components::spinner;
use crate::utils::{env, system};

pub fn run(args: StartArgs) {
    println!("Starting Runtipi...\n");

    let spin = spinner::new("");

    // User permissions
    spin.set_message("Checking user permissions");

    if let Err(_) = system::ensure_docker() {
        spin.fail("Your user is not allowed to run docker commands. Please add your user to the docker group or run the CLI as root.");
        spin.finish();
        return;
    }

    if let Err(_) = system::ensure_docker_compose_plugin() {
        spin.fail("Docker compose plugin is not installed. See https://docs.docker.com/compose/install/linux/ for more information");
        spin.finish();
        return;
    }

    spin.succeed("User permissions are ok");

    // Env file generation
    spin.set_message("Generating .env file...");

    if let Err(e) = env::generate_env_file(args.env_file) {
        spin.fail("Failed to generate .env file");
        spin.finish();
        eprintln!("Error: {}", e);
        return;
    }

    spin.succeed("Generated .env file");

    spin.set_message("Pulling images...");
    spin.finish();
}
