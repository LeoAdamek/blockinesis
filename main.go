package main

import (
	"github.com/spf13/viper"
	"github.com/spf13/pflag"
	"os"
	"github.com/sirupsen/logrus"
	"blockinesis/blockchain"
	"encoding/json"
	"time"
)

func main() {

	getConfig()

	c, err := blockchain.New()

	if err != nil {
		logrus.WithError(err).Fatalln("Unable to connect to Blockchain")
	}

	logfile, err := os.OpenFile("transactions.json", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	defer logfile.Close()

	if err != nil {
		logrus.Fatalln("Unable to open log file")
	}

	transactions := make(chan blockchain.Transaction)
	errors := make(chan error)

	go c.WatchTransactions(transactions, errors)




	go func() {
		for {
			err := <- errors

			logrus.WithError(err).Errorln("Error")
		}
	}()

	if until := viper.GetDuration("until"); until > 0 {
		go time.AfterFunc(until, func() {
			os.Exit(0)
		})
	}

	for {
		tx := <-transactions

		data, err := json.Marshal(tx)

		if err == nil {
			logfile.Write(data)
			logfile.Write([]byte("\n"))
		}
	}


}


func getConfig() {

	viper.AutomaticEnv()

	fs := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	fs.StringP("aws-region", "r", os.Getenv("AWS_REGION"), "AWS Region")
	fs.StringP("stream-name", "s", "blockinesis", "Kinesis Stream Name")
	fs.IntP("shard-count", "c", 2, "Shard Count")
	fs.StringP("ws-url", "u", "wss://ws.blockchain.info/inv", "WebSocket URL")
	fs.DurationP("until", "t", 0, "How long to wait before closing. (0 to continue forever)")

	fs.Parse(os.Args[1:])
	viper.BindPFlags(fs)

	err := viper.ReadInConfig()

	if err != nil {
		logrus.Warnln("Error Reading Config:", err)
	}

	logrus.Infoln("Configuration: ", viper.AllSettings())

}
