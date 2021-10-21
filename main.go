package main

import (
	"fmt"

	"os"

	"github.com/joho/godotenv"
	"riskVue.com/routers"
)

//Execution starts from main function
func main() {
	fmt.Print("aniket")
	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}
	//r := routers.SetupRouter()
	r := routers.SetupRouter()

	port := os.Getenv("port")

	//For run on requested port
	if len(os.Args) > 1 {
		reqPort := os.Args[1]
		if reqPort != "" {
			port = reqPort
		}
	}

	if port == "" {
		port = "8080" //localhost
	}
	type Job interface {
		Run()
	}

	r.Run(":" + port)

}
