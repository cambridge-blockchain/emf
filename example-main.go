package main

// This file is for demonstration purposes only to show how emf should be used to initialize a service

import (
	"os"

	emf "github.com/cambridge-blockchain/emf/emf"
	"github.com/cambridge-blockchain/emf/models"
	"github.com/cambridge-blockchain/emf/notifications"
)

func main() {
	configFile := os.Getenv("CONFIG")
	emfController := emf.New(configFile,
		models.BuildConfig{
			Version:          "v1.0.0",
			ReleaseTimestamp: "now",
			Component:        "xxx-component",
			Build:            "ef45432b78291",
		}, []notifications.NotificationType{},
	)

	emfController.GetLogger().Info("Server starting...")

	emfController.GetMiddlewares().UseMiddlewares(emfController.GetRouter())

	quit := emfController.GetServer().Startup()
	<-quit

	emfController.GetLogger().Info("Server dead, quitting...")
}
