package main

import (
	"github.com/spf13/viper"
	"github.com/spf13/pflag"
	"os"
	"github.com/sirupsen/logrus"
	"blockinesis/blockchain"
	"time"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
    "encoding/json"
)

func main() {

	getConfig()

	c, err := blockchain.New()

	if err != nil {
		logrus.WithError(err).Fatalln("Unable to connect to Blockchain")
	}

	transactions := make(chan blockchain.Transaction)
	errors := make(chan error)
	
	awsSession := session.Must(session.NewSession(&aws.Config{Region: aws.String(viper.GetString("aws-region"))}))
	k := kinesis.New(awsSession)

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
	
    t := time.NewTicker(viper.GetDuration("flush-interval"))
    
    var txns []blockchain.Transaction
    go func() {
        for tx := range transactions {
            txns = append(txns, tx)
        }
    }()
    
    for {
        <- t.C
        
        var messages []*kinesis.PutRecordsRequestEntry
        
        for _, tx := range txns {
            
            data, err := json.Marshal(tx)
    
            if err != nil {
                logrus.WithError(err).Warnln("Unable to encode Tx, skipping")
                continue
            }
            
            messages = append(messages, &kinesis.PutRecordsRequestEntry{
                Data: data,
                PartitionKey: &tx.Hash,
            })

        }
        
        txns = make([]blockchain.Transaction, 0)
        
        req := &kinesis.PutRecordsInput{
            StreamName: aws.String(viper.GetString("stream-name")),
            Records: messages,
        }
        
        result, err := k.PutRecords(req)
    
        if err != nil {
            logrus.WithError(err).Errorln("Failed to push data")
        } else {
            logrus.WithField("records", len(messages)).Infof("Pushed %d transactions.", len(messages))
        }
        
        if frc := *result.FailedRecordCount; frc > 0 {
            logrus.WithField("errors", frc).Warnf("Failed to push %d records.", frc)
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
	fs.DurationP("flush-interval", "f", 5*time.Second, "Flush interval")

	fs.Parse(os.Args[1:])
	viper.BindPFlags(fs)

	err := viper.ReadInConfig()

	if err != nil {
		logrus.Warnln("Error Reading Config:", err)
	}

	logrus.Infoln("Configuration: ", viper.AllSettings())

}
