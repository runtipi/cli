mod args;
mod commands;
mod components;
mod utils;

use args::RuntipiArgs;
use clap::Parser;

fn main() {
    let args = RuntipiArgs::parse();

    println!("Welcome to Runtipi CLI ✨\n");

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
            println!("{:?}", update_command);
        }
        args::RuntipiMainCommand::App(app_command) => {
            println!("{:?}", app_command);
        }
        args::RuntipiMainCommand::Debug => {
            commands::debug::run();
        }
    }
}
