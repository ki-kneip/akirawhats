package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type Status string

const (
	StatusConnecting   Status = "connecting"
	StatusQR           Status = "qr"
	StatusConnected    Status = "connected"
	StatusDisconnected Status = "disconnected"
	StatusLoggedOut    Status = "logged_out"
)

type Client struct {
	ID      string
	OwnerID string
	Inner   *whatsmeow.Client

	mu            sync.RWMutex
	qr            string
	status        Status
	phone         string
	webhookURL    string
	connectedCh   chan struct{}
	connectedOnce sync.Once
}

func NewClient(ctx context.Context, sessionID, webhookURL, ownerID string) (*Client, error) {
	waStore, err := getWAStore(ctx)
	if err != nil {
		return nil, err
	}
	deviceStore := waStore.NewDevice()
	inner := whatsmeow.NewClient(deviceStore, nil)

	c := &Client{
		ID:          sessionID,
		OwnerID:     ownerID,
		Inner:       inner,
		status:      StatusConnecting,
		webhookURL:  webhookURL,
		connectedCh: make(chan struct{}),
	}
	c.registerEvents()

	if inner.Store.ID == nil {
		qrChan, err := inner.GetQRChannel(ctx)
		if err != nil {
			return nil, fmt.Errorf("get qr channel: %w", err)
		}

		c.mu.Lock()
		c.status = StatusQR
		c.mu.Unlock()

		go func() {
			for item := range qrChan {
				switch item.Event {
				case "code":
					c.mu.Lock()
					c.qr = item.Code
					c.mu.Unlock()
				case "success":
					c.mu.Lock()
					c.status = StatusConnected
					c.qr = ""
					if inner.Store.ID != nil {
						c.phone = inner.Store.ID.User
					}
					c.mu.Unlock()
					c.connectedOnce.Do(func() { close(c.connectedCh) })
				case "timeout":
					c.mu.Lock()
					c.status = StatusDisconnected
					c.qr = ""
					c.mu.Unlock()
					c.connectedOnce.Do(func() { close(c.connectedCh) })
				}
			}
		}()

		if err := inner.Connect(); err != nil {
			return nil, fmt.Errorf("connect: %w", err)
		}
	} else {
		if err := inner.Connect(); err != nil {
			return nil, fmt.Errorf("connect: %w", err)
		}
		c.mu.Lock()
		c.status = StatusConnected
		if inner.Store.ID != nil {
			c.phone = inner.Store.ID.User
		}
		c.mu.Unlock()
		c.connectedOnce.Do(func() { close(c.connectedCh) })
	}

	return c, nil
}

func ReconnectClient(ctx context.Context, sessionID string, jid types.JID, webhookURL, ownerID string) (*Client, error) {
	waStore, err := getWAStore(ctx)
	if err != nil {
		return nil, err
	}
	deviceStore, err := waStore.GetDevice(ctx, jid)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	if deviceStore == nil {
		return nil, fmt.Errorf("device not found for jid %s", jid)
	}

	inner := whatsmeow.NewClient(deviceStore, nil)
	c := &Client{
		ID:          sessionID,
		OwnerID:     ownerID,
		Inner:       inner,
		status:      StatusConnecting,
		webhookURL:  webhookURL,
		connectedCh: make(chan struct{}),
	}
	c.registerEvents()

	if err := inner.Connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	c.mu.Lock()
	c.status = StatusConnected
	if inner.Store.ID != nil {
		c.phone = inner.Store.ID.User
	}
	c.mu.Unlock()
	c.connectedOnce.Do(func() { close(c.connectedCh) })

	return c, nil
}

func (c *Client) GetOwnerID() string {
	return c.OwnerID // immutable after creation, no lock needed
}

func (c *Client) GetStatus() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *Client) GetQR() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.qr
}

func (c *Client) GetPhone() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.phone
}

func (c *Client) GetWebhookURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.webhookURL
}

func (c *Client) SetWebhookURL(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.webhookURL = url
}

func (c *Client) GetJID() *types.JID {
	if c.Inner.Store.ID == nil {
		return nil
	}
	jid := *c.Inner.Store.ID
	return &jid
}

func (c *Client) SendText(ctx context.Context, to, text string) (whatsmeow.SendResponse, error) {
	jid, err := types.ParseJID(to)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("invalid JID %q: %w", to, err)
	}
	msg := &waProto.Message{
		Conversation: proto.String(text),
	}
	return c.Inner.SendMessage(ctx, jid, msg)
}

func (c *Client) Disconnect() {
	c.Inner.Disconnect()
	c.mu.Lock()
	c.status = StatusDisconnected
	c.mu.Unlock()
}

func (c *Client) Logout(ctx context.Context) error {
	err := c.Inner.Logout(ctx)
	c.mu.Lock()
	c.status = StatusLoggedOut
	c.mu.Unlock()
	return err
}

func (c *Client) registerEvents() {
	c.Inner.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			go c.deliverWebhook(v)
			body := v.Message.GetConversation()
			if body == "" {
				body = v.Message.GetExtendedTextMessage().GetText()
			}
			if body != "" {
				go persistMessage(c.ID, c.OwnerID, v.Info.Sender.String(), body, v.Info.Timestamp)
			}

		case *events.Disconnected:
			log.Printf("[%s] disconnected", c.ID)
			c.mu.Lock()
			c.status = StatusDisconnected
			c.mu.Unlock()
			go func() {
				if err := c.Inner.Connect(); err != nil {
					log.Printf("[%s] reconnect error: %v", c.ID, err)
					return
				}
				c.mu.Lock()
				c.status = StatusConnected
				c.mu.Unlock()
				log.Printf("[%s] reconnected", c.ID)
			}()

		case *events.LoggedOut:
			log.Printf("[%s] logged out: %v", c.ID, v.Reason)
			c.mu.Lock()
			c.status = StatusLoggedOut
			c.mu.Unlock()
		}
	})
}

type webhookPayload struct {
	Instance  string    `json:"instance"`
	From      string    `json:"from"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *Client) deliverWebhook(evt *events.Message) {
	url := c.GetWebhookURL()
	if url == "" {
		return
	}
	payload := webhookPayload{
		Instance:  c.ID,
		From:      evt.Info.Sender.String(),
		Message:   evt.Message.GetConversation(),
		Timestamp: evt.Info.Timestamp,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[%s] webhook marshal: %v", c.ID, err)
		return
	}

	const maxAttempts = 3
	backoff := time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.sendWebhook(url, b); err == nil {
			return
		} else if attempt < maxAttempts {
			log.Printf("[%s] webhook tentativa %d falhou: %v — retry em %s", c.ID, attempt, err, backoff)
			time.Sleep(backoff)
			backoff *= 2
		} else {
			log.Printf("[%s] webhook falhou após %d tentativas: %v", c.ID, maxAttempts, err)
		}
	}
}

func (c *Client) sendWebhook(url string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}
	return nil
}
