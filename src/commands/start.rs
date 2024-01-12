use crate::args::StartArgs;
use crate::utils::env;

pub fn run(_: StartArgs) {
    env::generate_env_file()
}
