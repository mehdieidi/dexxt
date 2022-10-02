package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	TELEGRAM_API_BASE_URL     = "https://api.telegram.org/bot"
	TELEGRAM_API_SEND_MESSAGE = "/sendMessage"
	BOT_TOKEN_ENV             = "TELEGRAM_BOT_TOKEN"
)

var telegramAPI string = TELEGRAM_API_BASE_URL + os.Getenv(BOT_TOKEN_ENV) + TELEGRAM_API_SEND_MESSAGE

// Update is a Telegram object that we receive every time a user interacts with the bot.
type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

// String implements the fmt.String interface to get the representation of an Update as a string.
func (u Update) String() string {
	return fmt.Sprintf("(update id: %d, message: %s)", u.UpdateId, u.Message)
}

// Message is a Telegram object that can be found in an update.
type Message struct {
	Text     string   `json:"text"`
	Chat     Chat     `json:"chat"`
	Audio    Audio    `json:"audio"`
	Voice    Voice    `json:"voice"`
	Document Document `json:"document"`
}

// String implements the fmt.String interface to get the representation of a Message as a string.
func (m Message) String() string {
	return fmt.Sprintf("(text: %s, chat: %s, audio %s)", m.Text, m.Chat, m.Audio)
}

// Audio refer to a audio file sent.
type Audio struct {
	FileId   string `json:"file_id"`
	Duration int    `json:"duration"`
}

// String implements the fmt.String interface to get the representation of an Audio as a string.
func (a Audio) String() string {
	return fmt.Sprintf("(file id: %s, duration: %d)", a.FileId, a.Duration)
}

// Voice can be summarized with similar attribute as an Audio message for our use case.
type Voice Audio

// Document refer to a file sent.
type Document struct {
	FileId   string `json:"file_id"`
	FileName string `json:"file_name"`
}

// String implements the fmt.String interface to get the representation of an Document as a string.
func (d Document) String() string {
	return fmt.Sprintf("(file id: %s, file name: %s)", d.FileId, d.FileName)
}

// Chat indicates the conversation to which the Message belongs.
type Chat struct {
	ID int `json:"id"`
}

// String implements the fmt.String interface to get the representation of a Chat as a string.
func (c Chat) String() string {
	return fmt.Sprintf("(id: %d)", c.ID)
}

// Handler sends a message back to the chat.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Parse incoming request
	update, err := parseIncomingRequest(r)
	if err != nil {
		log.Printf("error parsing incoming update, %s", err.Error())
		return
	}

	telegramResponseBody, err := sendToClient(update.Message.Chat.ID, strings.ToLower(update.Message.Text))
	if err != nil {
		log.Printf("got error %s from telegram, response body is %s", err.Error(), telegramResponseBody)
	} else {
		log.Printf("successfully distributed to chat id %d", update.Message.Chat.ID)
	}
}

// parseIncomingRequest parses incoming update to Update.
func parseIncomingRequest(r *http.Request) (*Update, error) {
	var update Update

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}

	if update.UpdateId == 0 {
		log.Printf("invalid update id, got update id = 0")
		return nil, errors.New("invalid update id of 0 indicates failure to parse incoming update")
	}

	return &update, nil
}

// sendToClient sends a text message to the Telegram chat identified by the chat ID.
func sendToClient(chatID int, incomingText string) (string, error) {
	if incomingText == "/start" {
		return "", nil
	}

	text, err := getFarsiAPI(incomingText)
	if err != nil {
		return "", err
	}

	log.Printf("Sending %s to chat_id: %d", text, chatID)

	response, err := http.PostForm(telegramAPI, url.Values{
		"chat_id": {strconv.Itoa(chatID)},
		"text":    {text},
	})
	if err != nil {
		log.Printf("error when posting text to the chat: %s", err.Error())
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("error in parsing telegram answer %s", err.Error())
		return "", err
	}

	log.Printf("Body of Telegram Response: %s", string(body))

	return string(body), nil
}

// getFarsi constructs and returns appropriate Farsi string associated with the given Finglish.
func getFarsi(finglish string) string {
	var farsi string

	for i := 0; i < len(finglish); i++ {
		switch finglish[i] {
		case 'a':
			farsi += "ا"
		case 'b':
			farsi += "ب"
		case 'c':
			if peekChar(i, finglish) == "h" {
				farsi += "چ"
				i++
			} else {
				farsi += "س"
			}
		case 'd':
			farsi += "د"
		case 'e':
			continue
		case 'f':
			farsi += "ف"
		case 'g':
			if peekChar(i, finglish) == "h" {
				farsi += "غ"
				i++
			} else {
				farsi += "گ"
			}
		case 'h':
			farsi += "ه"
		case 'i':
			farsi += "ی"
		case 'j':
			farsi += "ج"
		case 'k':
			if peekChar(i, finglish) == "h" {
				farsi += "خ"
				i++
			} else {
				farsi += "ک"
			}
		case 'l':
			farsi += "ل"
		case 'm':
			farsi += "م"
		case 'n':
			farsi += "ن"
		case 'o':
			farsi += "و"
		case 'p':
			farsi += "پ"
		case 'q':
			farsi += "ک"
		case 'r':
			farsi += "ر"
		case 's':
			if peekChar(i, finglish) == "h" {
				farsi += "ش"
				i++
			} else {
				farsi += "س"
			}
		case 't':
			farsi += "ت"
		case 'u':
			farsi += "و"
		case 'v':
			farsi += "و"
		case 'w':
			farsi += "و"
		case 'x':
			farsi += "خ"
		case 'y':
			farsi += "ی"
		case 'z':
			farsi += "ز"
		default:
			farsi += string(finglish[i])
		}
	}

	return farsi
}

// peekChar returns the next char in the given string if exists.
func peekChar(index int, str string) string {
	if index+1 < len(str) {
		return string(str[index+1])
	}
	return ""
}
