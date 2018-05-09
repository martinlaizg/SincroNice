package main

import (
	"log"
	"net/smtp"
)

func main() {

	from := "sincronicesl@gmail.com"
	pass := "Sincr0nice"
	to := "martinlaizg@gmail.com"
	subject := "Prueba"
	body := "Esto es una prueba"

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
