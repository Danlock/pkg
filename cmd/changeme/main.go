// Package main is an example entrypoint stub.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	// "github.com/joho/godotenv"
)

var (
	buildInfo = "NO INFO"
	buildTag  = "NO TAG"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	log.SetPrefix(buildTag + " ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	go func() { <-ctx.Done(); log.Fatal("Received signal, shutting down...") }()

	log.Printf("%s Built With: %s", buildInfo, runtime.Version())

	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		dotenvLocation string
	)

	// Example of using gotdotenv. Don't want to include this in this package's dependencies however.
	// if err := godotenv.Overload(dotenvLocation); err != nil {
	// 	log.Printf("No .env file found")
	// }
	flag.StringVar(&dotenvLocation, "e", "./ops/.env", "Location of .env file with environment variables in KEY=VALUE format. .env file takes precedence over real env vars.")
	flag.Parse()
}
