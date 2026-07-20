package main

import (
	"fmt"
	"time"

	"emailn/cmd/initializers"
	"emailn/internal/domain/campaign"
	"emailn/internal/infrastructure/database"
	"emailn/internal/infrastructure/mail"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {
	fmt.Println("started worker")
	db := database.NewDB()
	repository := database.CampaignRepository{Db: db}
	campaignService := campaign.ServiceImp{
		Repository: &repository,
		SendMail:   mail.SendMail,
	}

	for {
		campaigns, err := repository.GetCampaignsToBeSend()
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("Amount of campaigns: ", len(campaigns))

		for _, campaign := range campaigns {
			campaignService.SendEmailAndUpdateStatus(&campaign)
			fmt.Println("Campaign sent: ", campaign.ID)
		}

		// definition with base you need
		time.Sleep(10 * time.Second)
	}
}
