mod args;
mod commands;
mod components;
mod utils;

use args::RuntipiArgs;
use clap::Parser;
use colored::Colorize;

use crate::commands::update::UpdateArgs;
use crate::utils::env::env_string_to_map;

fn main() {
    let args = RuntipiArgs::parse();

    println!("{}", "Welcome to Runtipi CLI ✨\n".green());

    let current_dir = std::env::current_dir().unwrap_or_default();
    let env_file_path = current_dir.join(".env");
    let env_file = std::fs::read_to_string(&env_file_path).unwrap_or_default();
    let env_map = env_string_to_map(env_file.as_str());

    match args.command {
        args::RuntipiMainCommand::Start(args) => {
            commands::start::run(args);
        }
        args::RuntipiMainCommand::Stop => {
            commands::stop::run();
        }
        args::RuntipiMainCommand::Restart(args) => {
            commands::stop::run();
            commands::start::run(args);
        }
        args::RuntipiMainCommand::Update(update_command) => {
            let args = UpdateArgs {
                version: update_command.version.to_string(),
                env_file: update_command.env_file,
                no_permissions: update_command.no_permissions,
            };

            commands::stop::run();
            commands::update::run(args);
        }
        args::RuntipiMainCommand::ResetPassword => {
            commands::reset_password::run();
        }
        args::RuntipiMainCommand::App(app_command) => {
            commands::app::run(app_command, env_map);
        }
        args::RuntipiMainCommand::Debug => {
            commands::debug::run(env_map);
        }
    }
}
