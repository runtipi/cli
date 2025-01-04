use std::path::PathBuf;

use clap::{Args, Parser, Subcommand};

use crate::structs::{urn::Urn, version::VersionEnum};

#[derive(Debug, Subcommand)]
pub enum RuntipiMainCommand {
    /// Start your runtipi instance
    Start(StartArgs),
    /// Stop your runtipi instance
    Stop,
    /// Restart your runtipi instance
    Restart(StartArgs),
    /// Update your runtipi instance
    Update(UpdateCommand),
    /// Manage your apps
    App(AppCommand),
    /// Initiate a password reset for the admin user
    ResetPassword,
    /// Debug your runtipi instance
    Debug,
}

#[derive(Debug, Parser)]
#[clap(author, version, about)]
pub struct RuntipiArgs {
    #[clap(subcommand)]
    pub command: RuntipiMainCommand,
}

#[derive(Parser, Debug)]
pub struct StartArgs {
    /// Path to a custom .env file. Can be relative to the current directory or absolute.
    #[clap(short, long)]
    pub env_file: Option<PathBuf>,
    /// Skip setting file permissions (not recommended)
    #[clap(long)]
    pub no_permissions: bool,
}

#[derive(Debug, Args)]
pub struct UpdateCommand {
    /// The version to update to eg: v2.5.0 or latest
    pub version: VersionEnum,
    /// Path to a custom .env file. Can be relative to the current directory or absolute.
    #[clap(short, long)]
    pub env_file: Option<PathBuf>,
    /// Skip setting file permissions (not recommended)
    #[clap(long)]
    pub no_permissions: bool,
}

#[derive(Debug, Args)]
pub struct AppCommand {
    /// The subcommand to run
    #[clap(subcommand)]
    pub subcommand: AppSubcommand,
}

#[derive(Debug, Subcommand)]
pub enum AppSubcommand {
    /// Start an app
    Start(StartApp),
    /// Stop an app
    Stop(StopApp),
    /// Uninstall an app
    Uninstall(UninstallApp),
    /// Reset an app
    Reset(ResetApp),
    /// Update an app
    Update(UpdateApp),
    /// Start all apps
    StartAll(StartAll),
}

#[derive(Debug, Args)]
pub struct StartApp {
    /// The urn of the app to start
    pub urn: Urn,
}

#[derive(Debug, Args)]
pub struct StopApp {
    /// The urn of the app to stop
    pub urn: Urn,
}

#[derive(Debug, Args)]
pub struct UninstallApp {
    /// The urn of the app to uninstall
    pub urn: Urn,
}

#[derive(Debug, Args)]
pub struct ResetApp {
    /// The urn of the app to reset
    pub urn: Urn,
}

#[derive(Debug, Args)]
pub struct UpdateApp {
    /// The urn of the app to update
    pub urn: Urn,
}

#[derive(Debug, Args)]
pub struct StartAll {}
