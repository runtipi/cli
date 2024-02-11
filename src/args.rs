use semver::{Error as SemverError, Version};
use std::{path::PathBuf, str::FromStr};

use clap::{Args, Parser, Subcommand};

#[derive(Debug, Clone)]
pub enum VersionEnum {
    Version(Version),
    Latest,
    Nightly,
}

impl FromStr for VersionEnum {
    type Err = SemverError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        if s == "latest" {
            Ok(VersionEnum::Latest)
        } else if s == "nightly" {
            Ok(VersionEnum::Nightly)
        } else {
            // Remove the 'v' prefix if present
            let version_str = if s.starts_with('v') || s.starts_with('V') { &s[1..] } else { s };

            let version = Version::parse(version_str)?;
            Ok(VersionEnum::Version(version))
        }
    }
}

impl ToString for VersionEnum {
    fn to_string(&self) -> String {
        match self {
            VersionEnum::Version(version) => version.to_string(),
            VersionEnum::Latest => "latest".to_string(),
            VersionEnum::Nightly => "nightly".to_string(),
        }
    }
}

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
    /// The id of the app to start
    pub id: String,
}

#[derive(Debug, Args)]
pub struct StopApp {
    /// The id of the app to stop
    pub id: String,
}

#[derive(Debug, Args)]
pub struct UninstallApp {
    /// The id of the app to uninstall
    pub id: String,
}

#[derive(Debug, Args)]
pub struct ResetApp {
    /// The id of the app to reset
    pub id: String,
}

#[derive(Debug, Args)]
pub struct UpdateApp {
    /// The id of the app to update
    pub id: String,
}

#[derive(Debug, Args)]
pub struct StartAll {}
