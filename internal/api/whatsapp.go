package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"kikneip.com/akirawhats/internal/whatsapp"
)

type createInstanceReq struct {
	ID         string `json:"id" binding:"required"`
	WebhookURL string `json:"webhookUrl"`
}

type sendTextReq struct {
	To      string `json:"to" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type setWebhookReq struct {
	URL string `json:"url" binding:"required"`
}

func registrarWhatsApp(e gin.IRouter, sm *whatsapp.SessionManager) {
	grp := e.Group("/instance")

	grp.POST("", func(c *gin.Context) {
		var req createInstanceReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client, err := sm.Create(c.Request.Context(), req.ID, req.WebhookURL, getUserID(c))
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"id":         client.ID,
			"status":     client.GetStatus(),
			"qr":         client.GetQR(),
			"webhookUrl": client.GetWebhookURL(),
		})
	})

	grp.GET("", func(c *gin.Context) {
		clients := sm.ListByOwner(getUserID(c))
		out := make([]gin.H, 0, len(clients))
		for _, cl := range clients {
			out = append(out, gin.H{
				"id":     cl.ID,
				"status": cl.GetStatus(),
				"phone":  cl.GetPhone(),
			})
		}
		c.JSON(http.StatusOK, out)
	})

	grp.GET("/:id", func(c *gin.Context) {
		client, ok := sm.GetByOwner(c.Param("id"), getUserID(c))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":         client.ID,
			"status":     client.GetStatus(),
			"qr":         client.GetQR(),
			"phone":      client.GetPhone(),
			"webhookUrl": client.GetWebhookURL(),
		})
	})

	grp.GET("/:id/qr", func(c *gin.Context) {
		client, ok := sm.GetByOwner(c.Param("id"), getUserID(c))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		qr := client.GetQR()
		if qr == "" {
			c.JSON(http.StatusNoContent, nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"qr": qr})
	})

	grp.DELETE("/:id", func(c *gin.Context) {
		if err := sm.DeleteByOwner(c.Request.Context(), c.Param("id"), getUserID(c)); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "instance removed"})
	})

	grp.POST("/:id/webhook", func(c *gin.Context) {
		client, ok := sm.GetByOwner(c.Param("id"), getUserID(c))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		var req setWebhookReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client.SetWebhookURL(req.URL)
		sm.PersistWebhook(c.Request.Context(), client)
		c.JSON(http.StatusOK, gin.H{"webhookUrl": req.URL})
	})

	grp.GET("/:id/messages", func(c *gin.Context) {
		client, ok := sm.GetByOwner(c.Param("id"), getUserID(c))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		msgs, err := whatsapp.ListMessages(ctx, client.ID, getUserID(c), 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if msgs == nil {
			msgs = []whatsapp.MsgDoc{}
		}
		c.JSON(http.StatusOK, msgs)
	})

	grp.POST("/:id/send/text", func(c *gin.Context) {
		client, ok := sm.GetByOwner(c.Param("id"), getUserID(c))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		if client.GetStatus() != whatsapp.StatusConnected {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "instance not connected"})
			return
		}
		var req sendTextReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := client.SendText(c.Request.Context(), req.To, req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":        resp.ID,
			"timestamp": resp.Timestamp,
		})
	})
}
