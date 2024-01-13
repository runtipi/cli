use serde::{Deserialize, Serialize};

#[derive(Deserialize, Serialize, Debug)]
#[serde(untagged)]
pub enum StringOrInt {
    Str(String),
    Int(i64),
}

impl From<String> for StringOrInt {
    fn from(s: String) -> Self {
        StringOrInt::Str(s)
    }
}

impl From<&str> for StringOrInt {
    fn from(s: &str) -> Self {
        StringOrInt::Str(s.to_owned())
    }
}

impl StringOrInt {
    pub fn as_string(&self) -> String {
        match self {
            StringOrInt::Str(s) => s.clone(),
            StringOrInt::Int(i) => i.to_string(),
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SettingsSchema {
    #[serde(rename = "listenIp")]
    pub internal_ip: Option<String>,

    #[serde(rename = "nginxPort")]
    pub nginx_port: Option<StringOrInt>,

    #[serde(rename = "nginxPortSsl")]
    pub nginx_ssl_port: Option<StringOrInt>,

    #[serde(rename = "storagePath")]
    pub storage_path: Option<String>,

    #[serde(rename = "postgresPort")]
    pub postgres_port: Option<StringOrInt>,

    pub domain: Option<String>,

    #[serde(rename = "localDomain")]
    pub local_domain: Option<String>,
}
