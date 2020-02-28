package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"../go/pkg/amqpclient"
	"../go/pkg/containerutils"
	"../go/pkg/grpcendpoint"
)

func main() {
	if os.Getenv("HOME") != "/home/aasuser" {
		err := godotenv.Load("../../.env")
		if err == nil {
			log.Warn().Msg("***** DEVELOPMENT MODE: Successfully loaded .env file from repository root! *****")
		}
	}

	// Configure logging TimeField and line numbers
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	if os.Getenv("LOG_OUTPUT") == "CONSOLE" {
		output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
		output.FormatMessage = func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		}
		output.FormatFieldName = func(i interface{}) string {
			return fmt.Sprintf("%s:", i)
		}
		output.FormatFieldValue = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%s", i))
		}

		log.Logger = log.Output(output)
	}

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		log.Debug().Msg("LOG_LEVEL has been set to DEBUG")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var (
		AMQPCfg       amqpclient.Config
		GRPCEgressCfg grpcendpoint.GRPCEgressConfig
		GRPCEgress    grpcendpoint.GRPCEgress
	)

	amqpPort, _ := strconv.Atoi(os.Getenv("RABBITMQ_PORT"))
	AMQPCfg = amqpclient.Config{
		Host:     os.Getenv("RABBITMQ_HOST"),
		Port:     amqpPort,
		User:     os.Getenv("RABBITMQ_EGRESS_USER"),
		Password: os.Getenv("RABBITMQ_EGRESS_PASSWORD"),
		Exchange: os.Getenv("RABBITMQ_EGRESS_EXCHANGE"),
	}

	GRPCEgressCfg = grpcendpoint.GRPCEgressConfig{
		AMQPConfig: AMQPCfg,
	}

	services := []string{fmt.Sprintf("%s:%s", AMQPCfg.Host, strconv.Itoa(amqpPort))}
	containerutils.WaitForServices(services, time.Duration(60)*time.Second)

	GRPCEgress = grpcendpoint.NewGRPCEgress(GRPCEgressCfg)

	GRPCEgress.Init()

	containerutils.WaitForShutdown()

	GRPCEgress.Shutdown()
	os.Exit(0)
}