package control

import "strings"

type Commons struct {
	SellAll   bool      `json:"sellAll"`
	CancelAll bool      `json:"cancelAll"`
	Confirm   chan bool `json:"-"`
}

func (c Commons) String() string {
	var b strings.Builder
	if c.SellAll {
		b.WriteString("SellAll activated.\n")
	}
	if c.CancelAll {
		b.WriteString("CancelAll activated.\n")
	}

	return b.String()
}

func (c *Commons) Init() {
	c.Confirm = make(chan bool)
}
