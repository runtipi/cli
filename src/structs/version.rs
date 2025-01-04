use core::fmt;
use std::str::FromStr;

use semver::{Error as SemverError, Version};

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

impl fmt::Display for VersionEnum {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            VersionEnum::Version(version) => write!(f, "{}", version),
            VersionEnum::Latest => write!(f, "latest"),
            VersionEnum::Nightly => write!(f, "nightly"),
        }
    }
}
