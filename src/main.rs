mod args;
mod commands;
mod components;
mod utils;

use args::RuntipiArgs;
use clap::Parser;
use colored::Colorize;

use crate::commands::update::UpdateArgs;
use crate::utils::env::get_env_map;

fn main() {
    let args = RuntipiArgs::parse();

    println!("{}", "Welcome to Runtipi CLI ✨\n".green());

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
            let env_map = get_env_map();

            commands::app::run(app_command, env_map);
        }
        args::RuntipiMainCommand::Debug => {
            let env_map = get_env_map();

            commands::debug::run(env_map);
        }
    }
}
