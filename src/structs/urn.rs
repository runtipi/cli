use core::fmt;
use std::str::FromStr;

#[derive(Debug, Clone)]
pub struct Urn(String);

impl FromStr for Urn {
    type Err = &'static str;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        if input.split_once(':').is_some() {
            Ok(Self(input.to_string()))
        } else {
            Err("Must be <app-name>:<app-store-name>")
        }
    }
}

impl fmt::Display for Urn {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}
