package components

import (
	"github.com/Delta456/box-cli-maker/v2"
)

type ConsoleBox struct {
	title string
	body  string
	width int
	color string
}

func NewConsoleBox(title, body string, width int, boxColor string) *ConsoleBox {
	return &ConsoleBox{
		title: title,
		body:  body,
		width: width,
		color: boxColor,
	}
}

func (b *ConsoleBox) Print() {
	box := box.New(box.Config{
		Py:           2,
		Px:           2,
		Type:         "Double",
		Color:        "Green",
		TitlePos:     "Top",
		ContentAlign: "Center",
	})

	box.Print("Runtipi started successfully 🎉", b.body)
}
