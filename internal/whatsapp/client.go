package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	jid, err := c.parseRecipient(to)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}
	msg := &waProto.Message{Conversation: proto.String(text)}
	resp, err := c.Inner.SendMessage(ctx, jid, msg)
	if err == nil {
		go PersistSentMessage(c.ID, c.OwnerID, resp.ID, to, text, resp.Timestamp)
	}
	return resp, err
}

func (c *Client) SendImage(ctx context.Context, to string, data []byte, mimetype, caption string) (whatsmeow.SendResponse, error) {
	jid, err := c.parseRecipient(to)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}
	uploaded, err := c.Inner.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("upload image: %w", err)
	}
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimetype),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Caption:       proto.String(caption),
		},
	}
	resp, err := c.Inner.SendMessage(ctx, jid, msg)
	if err == nil {
		body := "[imagem]"
		if caption != "" {
			body = "[imagem] " + caption
		}
		go PersistSentMessage(c.ID, c.OwnerID, resp.ID, to, body, resp.Timestamp)
	}
	return resp, err
}

type GroupInfo struct {
	JID  string `json:"jid"`
	Name string `json:"name"`
}

func (c *Client) GetGroups(ctx context.Context) ([]GroupInfo, error) {
	groups, err := c.Inner.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("get groups: %w", err)
	}
	out := make([]GroupInfo, 0, len(groups))
	for _, g := range groups {
		out = append(out, GroupInfo{JID: g.JID.String(), Name: g.Name})
	}
	return out, nil
}

func (c *Client) parseRecipient(to string) (types.JID, error) {
	if !strings.Contains(to, "@") {
		to = to + "@s.whatsapp.net"
	}
	jid, err := types.ParseJID(to)
	if err != nil {
		return types.JID{}, fmt.Errorf("invalid JID %q: %w", to, err)
	}
	return jid, nil
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

		case *events.Receipt:
			status := MsgStatusDelivered
			if v.Type == events.ReceiptTypeRead {
				status = MsgStatusRead
			}
			for _, id := range v.MessageIDs {
				go UpdateMessageStatus(c.ID, id, status)
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
