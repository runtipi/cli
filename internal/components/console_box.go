package components

import (
	"fmt"
	"github.com/box-cli-maker/box-cli-maker/v3"
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
	box := box.NewBox().
		Style(box.Double).
		Padding(2, 2).
		TitlePosition(box.Top).
		ContentAlign(box.Center).
		Color(b.color)

	fmt.Println(box.MustRender(b.title, b.body))
}
