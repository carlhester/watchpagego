package utils

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
)

type Configuration struct {
	NotifyEmail string
	SenderEmail string
	Password    string
	Server      string
	Port        string
	Subject     string
}

func EmailResults(content string) {
	file, _ := os.Open("config.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}

	from := config.SenderEmail
	password := config.Password

	to := []string{
		config.NotifyEmail, config.NotifyEmail,
	}

	formattedcontent := "Sites with new content:\r\n\r\n" + content

	message := []byte("To: " + to[0] + "\r\n" +
		"Subject: " + config.Subject + "\r\n\r\n" +
		formattedcontent + "\r\n")

	auth := smtp.PlainAuth("", from, password, config.Server)

	err = smtp.SendMail(config.Server+":"+config.Port, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Email Sent!")

}
