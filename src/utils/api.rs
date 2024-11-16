use crate::utils::env::get_env_value;
use jsonwebtoken::{encode, Algorithm, EncodingKey, Header};

use reqwest::{
    blocking::{Body, Client, Response},
    Method,
};
use serde::{Deserialize, Serialize};
use std::io::{Error, ErrorKind};

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    sub: String,
}

pub fn create_client() -> Result<Client, Error> {
    let client = Client::builder().user_agent("reqwest").build();

    match client {
        Ok(c) => Ok(c),
        Err(err) => Err(Error::new(ErrorKind::Other, format!("Error creating client: {:?}", err))),
    }
}

fn create_token() -> Result<String, Error> {
    let claims = Claims { sub: "cli".to_string() };
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

pub fn api_request(url: String, method: Method, raw_body: &str) -> Result<Response, Error> {
    let client = create_client()?;
    let token = create_token()?;
    let auth_token = format!("Bearer {}", token);

    let body = Body::from(raw_body.to_string());

    let response = client
        .request(method, url)
        .header("Authorization", auth_token)
        .header("Content-Type", "application/json")
        .header("Accept", "application/json")
        .body(body)
        .send();

    match response {
        Ok(r) => Ok(r),
        Err(err) => Err(Error::new(ErrorKind::Other, format!("Error sending request: {:?}", err))),
    }
}
