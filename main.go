package main

import (
	"github.com/spf13/viper"
	"github.com/spf13/pflag"
	"os"
	"github.com/sirupsen/logrus"
	"blockinesis/blockchain"
	"time"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
    
    "github.com/guregu/dynamo"
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
	db := dynamo.New(awsSession)
	
	table := db.Table(viper.GetString("table-name"))

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

    // Collect the transactions into a slice for later batches.
    var txns []interface{}
    var total int64
    go func() {
        for tx := range transactions {
            txns = append(txns, tx)
            total += tx.Value
        }
    }()
    
    for {
        <- t.C
        
        written, err := table.Batch().Write().Put(txns...).Run()
    
        if err != nil {
            logrus.WithError(err).Errorln("Unable to push transactions")
            continue
        }
        
        txns = make([]interface{}, 0)
        
        totalBtc := float64(total)/1e9
        logrus.WithField("count", written).
            WithField("total", total).
                WithField("total_btc", totalBtc).
                Infof("Pushed %d transactions totalling %.6f BTC", written, totalBtc)
        
                total = 0
    }

}


func getConfig() {

	viper.AutomaticEnv()

	fs := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	fs.StringP("aws-region", "r", os.Getenv("AWS_REGION"), "AWS Region")
	fs.StringP("table-name", "t", "blockchain-transactions", "DynamoDB Table Name")
	fs.DurationP("flush-interval", "f", 5*time.Second, "Flush interval")

	fs.Parse(os.Args[1:])
	viper.BindPFlags(fs)

	err := viper.ReadInConfig()

	if err != nil {
		logrus.Warnln("Error Reading Config:", err)
	}

	logrus.Infoln("Configuration: ", viper.AllSettings())

}
