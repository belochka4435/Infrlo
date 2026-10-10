package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Токен берём из переменной окружения
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN не установлен")
	}

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Авторизован как @%s", bot.Self.UserName)

	// Long polling: получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Graceful shutdown: ловим SIGTERM (Infrlo шлёт его при остановке)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Бот запущен, жду сообщений...")

	// Основной цикл
	for {
		select {
		case update := <-updates:
			handleUpdate(bot, update)

		case <-stop:
			log.Println("Получен сигнал остановки, завершаю работу...")
			bot.StopReceivingUpdates()
			return
		}
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Игнорируем всё, кроме сообщений
	if update.Message == nil {
		return
	}

	msg := update.Message
	log.Printf("[%s] @%s: %s", msg.From.UserName, msg.From.FirstName, msg.Text)

	var reply string

	switch msg.Text {
	case "/start":
		reply = "👋 Привет! Я бот на Go, работаю на Infrlo через long polling.\nПопробуй: /ping, /help"
	case "/ping":
		reply = "🏓 pong!"
	case "/help":
		reply = "Доступные команды:\n/start — приветствие\n/ping — проверка связи\n/help — это сообщение"
	default:
		reply = "Ты написал: " + msg.Text
	}

	// Отправляем ответороюысдьыдлс
	_, err := bot.Send(tgbotapi.NewMessage(msg.Chat.ID, reply))
	if err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}
