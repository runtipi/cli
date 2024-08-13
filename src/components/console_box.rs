use colored::Colorize;
// Example of the Box component printed in the console:
//
// ╔════════════════════════ Tipi successfully started 🎉 ════════════════════════╗
// ║                                                                              ║
// ║             Visit: http://10.0.3.152:80 to access the dashboard              ║
// ║                                                                              ║
// ║             Find documentation and guides at: https://runtipi.io             ║
// ║                                                                              ║
// ║        Tipi is entirely written in TypeScript and we are looking for         ║
// ║                                contributors!                                 ║
// ║                                                                              ║
// ╚══════════════════════════════════════════════════════════════════════════════╝
//
// Example usage:
//
// let console_box = ConsoleBox::new(
//   "Runtipi started successfully",
//   "Find documentation and guides at: https://runtipi.io\n\nVisit: http://10.0.3.152:80 to access the dashboard\n\nTipi is entirely written in TypeScript and we are looking for contributors!",
//   80,
//   "green"
// );
// console_box.print();

#[derive(Debug)]
pub struct ConsoleBox {
    pub title: String,
    pub body: String,
    pub width: usize,
    pub color: String,
}

impl ConsoleBox {
    pub fn new(title: String, body: String, width: usize, color: String) -> ConsoleBox {
        ConsoleBox { title, body, width, color }
    }

    fn print_empty_line(&self) {
        println!(
            "{}{}{}",
            "║".color(self.color.as_ref()),
            " ".repeat(self.width - 2),
            "║".color(self.color.as_ref())
        );
    }

    pub fn print(&self) {
        // Find the longest line and set the box width
        let box_width = self.width;

        // Top border with a title
        let title = format!(" {} ", self.title);
        let top_border_len = box_width - title.len() - 2;
        let top_border_left = format!("╔{}", "═".repeat(top_border_len / 2));
        let top_border_right = format!("{}╗", "═".repeat(top_border_len / 2));

        // Print the top border
        println!(
            "{}{}{}",
            top_border_left.color(self.color.as_ref()),
            title.color(self.color.as_ref()),
            top_border_right.color(self.color.as_ref())
        );

        self.print_empty_line();
        // Print each line of the body, centered within the box
        for line in self.body.lines() {
            // If line is more than 80% of the box width, split it into multiple lines
            if line.len() > (box_width as f32 * 0.8) as usize {
                let mut split_lines = vec![];
                let mut current_line = String::new();
                for word in line.split_whitespace() {
                    if current_line.len() + word.len() > (box_width as f32 * 0.8) as usize {
                        split_lines.push(current_line);
                        current_line = String::new();
                    }
                    current_line.push_str(word);
                    current_line.push(' ');
                }
                split_lines.push(current_line);
                for line in split_lines {
                    let padded_line = format!("{:^width$}", line, width = box_width - 2);
                    println!("{}{}{}", "║".color(self.color.as_ref()), padded_line, "║".color(self.color.as_ref()));
                }
                continue;
            }

            let padded_line = format!("{:^width$}", line, width = box_width - 2);
            println!("{}{}{}", "║".color(self.color.as_ref()), padded_line, "║".color(self.color.as_ref()));
        }
        self.print_empty_line();

        // Bottom border
        let bottom_border = format!("╚{}╝", "═".repeat(box_width - 2));
        // Print the bottom border
        println!("{}", bottom_border.color(self.color.as_ref()));
    }
}
