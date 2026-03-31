package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"plugin/internal/utils"
	"time"
)

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type embedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type embedAuthor struct {
	Name    string `json:"name,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type embed struct {
	Author      *embedAuthor `json:"author,omitempty"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []embedField `json:"fields,omitempty"`
	Footer      *embedFooter `json:"footer,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
}

type payload struct {
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Content   string  `json:"content,omitempty"`
	Embeds    []embed `json:"embeds,omitempty"`
}

type Webhook struct {
	URL    string
	client *http.Client
}

func New(url string) *Webhook {
	return &Webhook{
		URL:    url,
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

func (w *Webhook) send(p payload) {
	if w.URL == "" {
		return
	}
	b, err := json.Marshal(p)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, w.URL, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := w.client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}

func basePayload() payload {
	return payload{
		Username:  "Hypno Bot",
		AvatarURL: "https://media.discordapp.net/attachments/1475891647190405192/1476325227247177934/PLUTOPLUHH.png?ex=69a0b682&is=699f6502&hm=1a6e0a5918f3bac9e1108ac08d5b6fa50fa4af0b3c6433ab07c31f40df54190e&=&format=webp&quality=lossless",
	}
}

func (w *Webhook) WinWebhook(player string, amount int) {
	p := basePayload()
	p.Embeds = []embed{
		{
			Author: &embedAuthor{
				Name: "🎰  Casino — Win",
			},
			Color: 0x2ECC71,
			Fields: []embedField{
				{Name: "Player", Value: "**" + player + "**", Inline: true},
				{Name: "Payout", Value: "**" + utils.FormatMoney(amount) + "**", Inline: true},
				{Name: "Result", Value: "**WIN**", Inline: true},
			},
			Footer:    &embedFooter{Text: "Gambling bot  •  Win Log"},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	w.send(p)
}

func (w *Webhook) LossWebhook(player string, amount int) {
	p := basePayload()
	p.Embeds = []embed{
		{
			Author: &embedAuthor{
				Name: "🎰  Casino — Loss",
			},
			Color: 0xE74C3C,
			Fields: []embedField{
				{Name: "Player", Value: "**" + player + "**", Inline: true},
				{Name: "Amount Lost", Value: "**" + utils.FormatMoney(amount) + "**", Inline: true},
				{Name: "Result", Value: "**LOSS**", Inline: true},
			},
			Footer:    &embedFooter{Text: "Gambling bot  •  Loss Log"},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	w.send(p)
}

func (w *Webhook) PayWebhook(sender, receiver string, amount int) {
	p := basePayload()
	p.Embeds = []embed{
		{
			Author: &embedAuthor{
				Name: "🎰  Casino — Payment",
			},
			Color: 0x9B59B6,
			Fields: []embedField{
				{Name: "Sender", Value: "**" + sender + "**", Inline: true},
				{Name: "Receiver", Value: "**" + receiver + "**", Inline: true},
				{Name: "Amount", Value: "**" + utils.FormatMoney(amount) + "**", Inline: true},
			},
			Footer:    &embedFooter{Text: "Casino Server  •  Transfer Log"},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	w.send(p)
}
