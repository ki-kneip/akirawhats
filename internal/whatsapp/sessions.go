package whatsapp

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mau.fi/whatsmeow/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"kikneip.com/akirawhats/internal/db"
)

const sessionsCollection = "wa_sessions"

type sessionDoc struct {
	ID         string `bson:"_id"`
	JIDString  string `bson:"jid"`
	Phone      string `bson:"phone"`
	WebhookURL string `bson:"webhook_url,omitempty"`
	OwnerID    string `bson:"owner_id"`
}

type SessionManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewSessionManager() *SessionManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SessionManager{
		clients: make(map[string]*Client),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (sm *SessionManager) RestoreSessions(ctx context.Context) {
	col := db.Collection(sessionsCollection)
	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("restore sessions: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var docs []sessionDoc
	if err := cursor.All(ctx, &docs); err != nil {
		log.Printf("decode sessions: %v", err)
		return
	}

	for _, rec := range docs {
		jid, err := types.ParseJID(rec.JIDString)
		if err != nil {
			log.Printf("parse jid %s: %v", rec.JIDString, err)
			continue
		}
		client, err := ReconnectClient(ctx, rec.ID, jid, rec.WebhookURL, rec.OwnerID)
		if err != nil {
			log.Printf("restore session %s: %v", rec.ID, err)
			continue
		}
		sm.mu.Lock()
		sm.clients[rec.ID] = client
		sm.mu.Unlock()
		log.Printf("restored session: %s (%s)", rec.ID, rec.Phone)
	}
}

func (sm *SessionManager) Create(ctx context.Context, sessionID, webhookURL, ownerID string) (*Client, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.clients[sessionID]; exists {
		return nil, fmt.Errorf("session %q already exists", sessionID)
	}

	client, err := NewClient(sm.ctx, sessionID, webhookURL, ownerID)
	if err != nil {
		return nil, err
	}

	sm.clients[sessionID] = client
	go sm.waitAndPersist(sm.ctx, client)
	return client, nil
}

func (sm *SessionManager) waitAndPersist(ctx context.Context, c *Client) {
	select {
	case <-c.connectedCh:
		if c.GetStatus() != StatusConnected {
			return
		}
		jid := c.GetJID()
		if jid == nil {
			return
		}
		sm.upsertSession(ctx, c, jid)
	case <-ctx.Done():
	}
}

func (sm *SessionManager) upsertSession(ctx context.Context, c *Client, jid *types.JID) {
	col := db.Collection(sessionsCollection)
	doc := sessionDoc{
		ID:         c.ID,
		JIDString:  jid.String(),
		Phone:      c.GetPhone(),
		WebhookURL: c.GetWebhookURL(),
		OwnerID:    c.OwnerID,
	}
	opts := options.Replace().SetUpsert(true)
	if _, err := col.ReplaceOne(ctx, bson.M{"_id": c.ID}, doc, opts); err != nil {
		log.Printf("persist session %s: %v", c.ID, err)
	}
}

func (sm *SessionManager) PersistWebhook(ctx context.Context, c *Client) {
	jid := c.GetJID()
	if jid == nil {
		return
	}
	sm.upsertSession(ctx, c, jid)
}

func (sm *SessionManager) Get(sessionID string) (*Client, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	c, ok := sm.clients[sessionID]
	return c, ok
}

func (sm *SessionManager) GetByOwner(sessionID, ownerID string) (*Client, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	c, ok := sm.clients[sessionID]
	if !ok || c.OwnerID != ownerID {
		return nil, false
	}
	return c, true
}

func (sm *SessionManager) ListByOwner(ownerID string) []*Client {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	var clients []*Client
	for _, c := range sm.clients {
		if c.OwnerID == ownerID {
			clients = append(clients, c)
		}
	}
	return clients
}

func (sm *SessionManager) Delete(ctx context.Context, sessionID string) error {
	sm.mu.Lock()
	c, ok := sm.clients[sessionID]
	if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("session %q not found", sessionID)
	}
	delete(sm.clients, sessionID)
	sm.mu.Unlock()

	// I/O happens outside the lock so reads (Get/List) are not blocked.
	if err := c.Logout(ctx); err != nil {
		log.Printf("logout session %s: %v", sessionID, err)
	}
	col := db.Collection(sessionsCollection)
	if _, err := col.DeleteOne(ctx, bson.M{"_id": sessionID}); err != nil {
		log.Printf("delete session record %s: %v", sessionID, err)
	}
	return nil
}

func (sm *SessionManager) DeleteByOwner(ctx context.Context, sessionID, ownerID string) error {
	sm.mu.Lock()
	c, ok := sm.clients[sessionID]
	if !ok || c.OwnerID != ownerID {
		sm.mu.Unlock()
		return fmt.Errorf("session %q not found", sessionID)
	}
	delete(sm.clients, sessionID)
	sm.mu.Unlock()

	if err := c.Logout(ctx); err != nil {
		log.Printf("logout session %s: %v", sessionID, err)
	}
	col := db.Collection(sessionsCollection)
	if _, err := col.DeleteOne(ctx, bson.M{"_id": sessionID}); err != nil {
		log.Printf("delete session record %s: %v", sessionID, err)
	}
	return nil
}

func (sm *SessionManager) Disconnect(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	c, ok := sm.clients[sessionID]
	if !ok {
		return fmt.Errorf("session %q not found", sessionID)
	}

	c.Disconnect()
	delete(sm.clients, sessionID)
	return nil
}

func (sm *SessionManager) DisconnectAll() {
	sm.cancel()
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for id, c := range sm.clients {
		c.Disconnect()
		delete(sm.clients, id)
	}
}

func (sm *SessionManager) List() []*Client {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	clients := make([]*Client, 0, len(sm.clients))
	for _, c := range sm.clients {
		clients = append(clients, c)
	}
	return clients
}
