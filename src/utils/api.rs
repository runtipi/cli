use crate::utils::env::get_env_value;
use jsonwebtoken::{encode, Algorithm, EncodingKey, Header};

use reqwest::blocking::{Client, Response};
use serde::{Deserialize, Serialize};
use std::io::{Error, ErrorKind};

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    sub: String,
}

fn create_client() -> Result<Client, Error> {
    let client = Client::builder().build();

    match client {
        Ok(c) => Ok(c),
        Err(err) => Err(Error::new(ErrorKind::Other, format!("Error creating client: {:?}", err))),
    }
}

fn create_token() -> Result<String, Error> {
    let claims = Claims { sub: "1".to_string() };
    let jwt_secret = get_env_value("JWT_SECRET");

    match jwt_secret {
        Some(secret) => {
            let encoding_key = EncodingKey::from_secret(secret.as_ref());
            let encoded = encode(&Header::new(Algorithm::HS256), &claims, &encoding_key);

            match encoded {
                Ok(t) => Ok(t),
                Err(err) => Err(Error::new(ErrorKind::Other, format!("Error creating token: {:?}", err))),
            }
        }
        None => Err(Error::new(ErrorKind::Other, "JWT_SECRET not found in environment variables")),
    }
}

pub fn api_request(url: String) -> Result<Response, Error> {
    let client = create_client()?;
    let token = create_token()?;
    let auth_token = format!("Bearer {}", token);
    let response = client.post(url).header("Authorization", auth_token).send();

    match response {
        Ok(r) => Ok(r),
        Err(err) => Err(Error::new(ErrorKind::Other, format!("Error sending request: {:?}", err))),
    }
}
