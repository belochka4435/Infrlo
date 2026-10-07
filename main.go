package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// ВАЖНО: логируем ВСЁ сразу, чтобы видеть в панели
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("🚀 Бот стартует...")

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ BOT_TOKEN не установлен в переменных окружения!")
	}
	log.Println("✅ Токен получен")

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Ошибка авторизации: %v", err)
	}
	bot.Debug = true // Включаем дебаг для первых запусков
	log.Printf("✅ Авторизован как @%s", bot.Self.UserName)

	// === HTTP health check ===
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Запускаем HTTP в отдельной горутине
	go func() {
		log.Printf("🌐 HTTP health check слушает порт %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	// ==========================

	// Long polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🤖 Бот запущен, жду сообщений...")

	for {
		select {
		case update := <-updates:
			handleUpdate(bot, update)
		case <-stop:
			log.Println("🛑 Получен сигнал остановки")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			httpServer.Shutdown(ctx)
			bot.StopReceivingUpdates()
			return
		}
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	log.Printf("📨 [%s] %s: %s", msg.From.UserName, msg.From.FirstName, msg.Text)

	var reply string
	switch msg.Text {
	case "/start":
		reply = "👋 Привет! Я работаю на Infrlo.\n/start, /ping, /help"
	case "/ping":
		reply = "🏓 pong!"
	case "/help":
		reply = "/start — приветствие\n/ping — проверка\n/help — помощь"
	default:
		reply = "Ты написал: " + msg.Text
	}

	if _, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, reply)); err != nil {
		log.Printf("❌ Ошибка отправки: %v", err)
	}
}
