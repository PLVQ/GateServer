package nsqer

import (
	"fmt"
	"gateServer/log"

	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

func InitConsuemr(topic string, channel string, handle nsq.Handler) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Log.WithField("Error", err.Error()).Fatal()
	}

	consumer.AddHandler(handle)
	if err = consumer.ConnectToNSQD("127.0.0.1:4150"); err != nil {
		fmt.Println(err)
	}

	<-consumer.StopChan
	log.Log.WithFields(logrus.Fields{"topic": topic, "channel": channel}).Error("stop nsq consume")
}
