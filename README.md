# Telebot
Telebot is an application that allows available multiple services to work with a single Telegram bot. Kafka is used for the communication. The application consists of 3 parts:
- The server works with a Telegram bot and receives data from services.
- The Kafka provides the data communication.
- The client is used by servers to send data to the server.

A bot has the commands that was sent by services. The server receives user data from the bot and sends it to the servers, which process the data and send replies.
