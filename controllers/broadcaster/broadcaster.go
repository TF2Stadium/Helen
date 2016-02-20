// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package broadcaster

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
	"github.com/streadway/amqp"
)

const (
	room = iota
	steamID
)

var queueName string

type message struct {
	TargetType int
	Target     string // can be a room or a steamid

	Event   string
	Content interface{}
}

func StartListening() {
	err := helpers.AMQPChannel.ExchangeDeclare(config.Constants.RabbitMQExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	q, err := helpers.AMQPChannel.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	queueName = q.Name

	err = helpers.AMQPChannel.QueueBind(q.Name, "", config.Constants.RabbitMQExchange, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	deliveries, err := helpers.AMQPChannel.Consume(q.Name, "", true, false, false, false, nil)
	go func() {
		for d := range deliveries {
			if d.CorrelationId == queueName {
				continue
			}

			var msg message
			json.Unmarshal(d.Body, &msg)

			switch msg.TargetType {
			case room:
				sendMessageToRoom(msg.Target, msg.Event, msg.Content)
			case steamID:
				sendMessage(msg.Target, msg.Event, msg.Content)
			}
		}
	}()

}

func publishMessage(targetType int, target, event string, content interface{}) {
	bytes, _ := json.Marshal(message{targetType, target, event, content})
	publish := amqp.Publishing{
		CorrelationId: queueName,
		ContentType:   "application/json",
		Body:          bytes,
	}

	err := helpers.AMQPChannel.Publish(config.Constants.RabbitMQExchange, "", false, false, publish)
	if err != nil {
		logrus.Error(err)
	}

}

func SendMessage(steamid string, event string, content interface{}) {
	found := sendMessage(steamid, event, content)
	if !found && helpers.AMQPChannel != nil {
		//fan out message
		publishMessage(steamID, steamid, event, content)
	}
}

func SendMessageToRoom(r string, event string, content interface{}) {
	sendMessageToRoom(r, event, content)
	//fan out message to exchange
	if helpers.AMQPChannel != nil {
		publishMessage(room, r, event, content)
	}
}

func sendMessage(steamid string, event string, content interface{}) bool {
	sockets, ok := sessions.GetSockets(steamid)
	if !ok {
		return false
	}

	for _, socket := range sockets {
		go func(so *wsevent.Client) {
			so.EmitJSON(helpers.NewRequest(event, content))
		}(socket)
	}

	return true
}

func sendMessageToRoom(room string, event string, content interface{}) {
	v := helpers.NewRequest(event, content)

	socket.AuthServer.BroadcastJSON(room, v)
	socket.UnauthServer.BroadcastJSON(room, v)
}
