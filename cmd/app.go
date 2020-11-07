package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"my5G-RANTester/internal/templates"
	"os"
	"strconv"
)

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "load-test",
				Aliases: []string{"load-test"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example: load-test -n 5 \n" +
					"Example for concurrency testing with different GNBs: load-test -g -n 10\n" +
					"Example for concurrency testing with some TNLAs: load-test -t -n 10\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-ues", Value: 1, Aliases: []string{"n"}},
					&cli.BoolFlag{Name: "gnb", Aliases: []string{"g"}},
					&cli.BoolFlag{Name: "tnla", Aliases: []string{"t"}},
				},
				Action: func(c *cli.Context) error {

					if c.String("tnla") == "true" {
						numUes, err := strconv.Atoi(c.String("number-of-ues"))
						if err != nil {
							fmt.Println("Error in convert string to int")
						}
						fmt.Printf("Testing attach with %d ues in TNLAs\n", numUes)
						fmt.Println(templates.TestMultiAttachUesInConcurrencyWithTNLAs(numUes))
					} else if c.String("gnb") == "true" {
						numUes, err := strconv.Atoi(c.String("number-of-ues"))
						if err != nil {
							fmt.Println("Error in convert string to int")
						}
						fmt.Printf("Testing attach with %d ues in different GNBs\n", numUes)
						fmt.Println(templates.TestMultiAttachUesInConcurrencyWithGNBs(numUes))
					} else {
						numUes, err := strconv.Atoi(c.String("number-of-ues"))
						if err != nil {
							fmt.Println("Error in convert string to int")
						}
						fmt.Printf("Testing attach with %d ues\n", numUes)
						fmt.Println(templates.TestMultiAttachUesInQueue(numUes))
					}
					return nil
				},
			},
			{
				Name:    "ue",
				Aliases: []string{"ue"},
				Usage:   "test attach for ue with configuration",
				Action: func(c *cli.Context) error {
					fmt.Println(templates.TestAttachUeWithConfiguration())
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
