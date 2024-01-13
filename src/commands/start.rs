use crate::args::StartArgs;
use crate::components::spinner;
use crate::utils::env;

pub fn run(args: StartArgs) {
    let spin = spinner::new("Generating .env file...");

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
