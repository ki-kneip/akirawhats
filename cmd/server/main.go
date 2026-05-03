package main

import (
	"context"
	stderrors "errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/api"
	"kikneip.com/akirawhats/internal/db"
	"kikneip.com/akirawhats/internal/model"
	"kikneip.com/akirawhats/internal/repo"
	"kikneip.com/akirawhats/internal/service"
	"kikneip.com/akirawhats/internal/whatsapp"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	validateEnv()

	ctx := context.Background()

	db.Connect(ctx)
	defer db.Disconnect(ctx)

	userRepo := &repo.UserImpl{}
	userSvc := service.NewUserService(userRepo)

	seedAdmin(ctx, userRepo, userSvc)

	sm := whatsapp.NewSessionManager()
	sm.RestoreSessions(ctx)

	engine := gin.Default()
	api.RegistrarRotas(engine, userSvc, sm)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !stderrors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error running http server: %v", err)
		}
	}()
	log.Printf("server running on :%s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sm.DisconnectAll()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func validateEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("[FATAL] JWT_SECRET não está configurado — abortando inicialização")
	}
	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Println("[WARN] JWT_SECRET tem menos de 32 caracteres — use um segredo mais longo em produção")
	}
	if os.Getenv("MONGO_URI") == "" {
		log.Fatal("[FATAL] MONGO_URI não está configurado — abortando inicialização")
	}
	if os.Getenv("ADMIN_EMAIL") == "" {
		log.Fatal("[FATAL] ADMIN_EMAIL não está configurado — abortando inicialização")
	}
	if os.Getenv("ADMIN_PASSWORD") == "" {
		log.Fatal("[FATAL] ADMIN_PASSWORD não está configurado — abortando inicialização")
	}
}

func seedAdmin(ctx context.Context, r internal.UserRepo, svc internal.UserService) {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")

	_, err := r.GetUserByEmail(ctx, email)
	if err == nil {
		return // já existe
	}
	if !stderrors.Is(err, internal.ErrNotFound) {
		log.Printf("[WARN] seed admin check failed: %v", err)
		return
	}

	_, createErr := svc.CreateUser(ctx, model.UserDTOPost{
		FirstName: "Admin",
		LastName:  "",
		Email:     email,
		Password:  password,
	})
	if createErr != nil {
		log.Printf("[WARN] seed admin create failed: %v", createErr)
		return
	}
	log.Printf("admin user created: %s", email)
}
