package telegram

import (
	"context"
	"fmt"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// PairResult holds sender info captured during a pairing handshake.
type PairResult struct {
	UserID    int64  `json:"telegram_user_id"`
	Username  string `json:"telegram_username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"message_text"`
}

// DisplayName returns "FirstName LastName" for display purposes.
func (p PairResult) DisplayName() string {
	name := p.FirstName
	if p.LastName != "" {
		name += " " + p.LastName
	}
	return name
}

// Pair connects to the Telegram Bot API and waits for the first non-bot
// message, bypassing the access gate. It is used during initial setup to
// capture the user's Telegram ID and chat ID without requiring them to
// look up these values manually.
//
// Security: The pairing window is bounded by timeout (caller should enforce
// a max of 300s). Only one message is captured; it is never relayed to
// Gas Town. If no message arrives or the caller cancels, no config changes
// are made (fail-closed).
func Pair(ctx context.Context, token string, timeout time.Duration) (PairResult, error) {
	client := &http.Client{Timeout: httpClientTimeout}
	api, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil {
		return PairResult{}, fmt.Errorf("telegram pair: connect: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = pollTimeout

	updates := api.GetUpdatesChan(u)
	defer api.StopReceivingUpdates()

	deadline := time.After(timeout)
	for {
		select {
		case <-ctx.Done():
			return PairResult{}, ctx.Err()
		case <-deadline:
			return PairResult{}, fmt.Errorf("no message received within %s", timeout)
		case update, ok := <-updates:
			if !ok {
				return PairResult{}, fmt.Errorf("telegram pair: update channel closed")
			}
			if update.Message == nil {
				continue
			}
			msg := update.Message
			if msg.From == nil || msg.From.IsBot {
				continue
			}
			return PairResult{
				UserID:    msg.From.ID,
				Username:  msg.From.UserName,
				FirstName: msg.From.FirstName,
				LastName:  msg.From.LastName,
				ChatID:    msg.Chat.ID,
				Text:      msg.Text,
			}, nil
		}
	}
}
