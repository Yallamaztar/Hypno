package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"plugin/internal/config"
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
	cfg    *config.Config
	client *http.Client
}

func New(url string, cfg *config.Config) *Webhook {
	return &Webhook{
		URL:    url,
		cfg:    cfg,
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
			Color: 0x00D084,
			Fields: []embedField{
				{Name: "Player", Value: "**" + player + "**", Inline: true},
				{Name: "Payout", Value: "**" + w.cfg.Gambling.Currency + utils.FormatMoney(amount) + "**", Inline: true},
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
			Color: 0xC0392B,
			Fields: []embedField{
				{Name: "Player", Value: "**" + player + "**", Inline: true},
				{Name: "Amount Lost", Value: "**" + w.cfg.Gambling.Currency + utils.FormatMoney(amount) + "**", Inline: true},
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
			Color: 0x3A86FF,
			Fields: []embedField{
				{Name: "Sender", Value: "**" + sender + "**", Inline: true},
				{Name: "Receiver", Value: "**" + receiver + "**", Inline: true},
				{Name: "Amount", Value: "**" + w.cfg.Gambling.Currency + utils.FormatMoney(amount) + "**", Inline: true},
			},
			Footer:    &embedFooter{Text: "Gambling bot  •  Transfer Log"},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	w.send(p)
}

func (w *Webhook) MaxBetWebhook(player string, amount int) {
	desc := "**Max bet has been disabled**"
	color := 0xF4A261

	if amount > 0 {
		desc = fmt.Sprintf("**Max bet set to %s%s**", w.cfg.Gambling.Currency, utils.FormatMoney(amount))
		color = 0xE9C46A
	}

	p := basePayload()
	p.Embeds = []embed{
		{
			Author: &embedAuthor{
				Name: "🎰  Casino — Max Bet",
			},
			Description: desc,
			Color:       color,
			Fields: []embedField{
				{
					Name:   "Player",
					Value:  "**" + player + "**",
					Inline: true,
				},
				{Name: "Status", Value: func() string {
					if amount > 0 {
						return "**🟢 ENABLED**"
					}
					return "**🔴 DISABLED**"
				}(),
					Inline: true,
				},
			},
			Footer: &embedFooter{
				Text: "Gambling bot  •  Max Bet Log",
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	w.send(p)
}
