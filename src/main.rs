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
        args::RuntipiMainCommand::Start(start_args) => {
            commands::start::run(start_args);
        }
        args::RuntipiMainCommand::Stop => {
            println!("Stop");
        }
        args::RuntipiMainCommand::Restart(restart_args) => {
            println!("{:?}", restart_args);
        }
        args::RuntipiMainCommand::Update(update_command) => {
            println!("{:?}", update_command);
        }
        args::RuntipiMainCommand::App(app_command) => {
            println!("{:?}", app_command);
        }
    }
}
