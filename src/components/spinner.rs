use std::time::Duration;

use indicatif::{ProgressBar, ProgressStyle};

pub struct CustomSpinner {
    spinner: ProgressBar,
}

impl CustomSpinner {
    fn new(initial_message: String) -> CustomSpinner {
        let spinner = ProgressBar::new_spinner();
        let style: &[&str] = &["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
        let progress_style = ProgressStyle::default_spinner();

        spinner.set_style(ProgressStyle::tick_strings(progress_style, style));

        spinner.set_message(initial_message);
        spinner.enable_steady_tick(Duration::from_millis(50));

        CustomSpinner { spinner }
    }

    pub fn succeed(&self, message: &str) {
        const CHECKMARK: &str = "\u{001b}[32;1m\u{2713}\u{001b}[0m";

        let success_message = format!("{} {}", CHECKMARK, message);

        self.spinner.println(success_message);
    }

    pub fn fail(&self, message: &str) {
        const CROSSMARK: &str = "\u{001b}[31;1m\u{2717}\u{001b}[0m";

        let failure_message = format!("{} {}", CROSSMARK, message);

        self.spinner.println(failure_message);
    }

    // pub fn warn(&self, message: &str) {
    //     const WARNING: &str = "\u{001b}[33;1m\u{26A0}\u{001b}[0m";

    //     let warning_message = format!("{} {}", WARNING, message);

    //     self.spinner.println(warning_message);
    // }

    pub fn set_message(&self, message: &str) {
        self.spinner.set_message(message.to_string());
    }

    pub fn finish(&self) {
        self.spinner.finish_and_clear();
    }
}

pub fn new(message: &str) -> CustomSpinner {
    CustomSpinner::new(message.to_string())
}
