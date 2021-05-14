package main

import (
	"net"
	"context"
	"fmt"
	"log"
	"encoding/json"

	"k8s.io/client-go/tools/cache"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/logging/logkey"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"

	v1alpha1 "github.com/cowbon/brokerchannel/pkg/apis/samples/v1alpha1"
	brokerchannelinformer "github.com/cowbon/brokerchannel/pkg/client/informers"
	"knative.dev/pkg/apis"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type MQTTConnection struct {
	client paho.Client
	logger *zap.SugaredLogger
	topic string
	addr *apis.URL
	ceClient	cloudevents.Client
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}


func newMQTTConnection(addr *apis.URL, topic, broker string, port int, logger *zap.SugaredLogger) (*MQTTConnection, error) {
	server := fmt.Sprintf("tcp://%s:%d", broker, port)
	logger.Infof("Create connection to %s", *server)
	conn, err := net.Dial("tcp", *server)
	if err != nil {
		return nil, err
	}

	client := paho.NewClient(paho.ClientConfig{
		Conn: conn,
	})
	client.Router = paho.newSingleHandlerRouter(func(m *paho.Publish) {
		event := cloudevents.NewEvent()
		for _, prop := range m.Properties.User {
			switch key := prop.Key {
			case "source":
				event.SetSource(prop.Value)
			case "type":
				event.SetType(prop.Value)
			case "ID":
				event.SetID(prop.Value)
			}
		}
		event.SetData(cloudevents.ApplicationJSON, m.Payload)
		ctx := cloudevents.ContextWithTarget(context.Background(), addr)
		if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
			logger.Fatalf("failed to send, %v", result)
		}
	})
	ca, err := c.Connect(context.Background(), &paho.Connect{CleanStart: true,})
	if err != nil {
		return nil, err
	}
	if ca.ReasonCode != 0 {
		logger.Errorf("Failed to connect to %s : %d - %s", *server, ca.ReasonCode, ca.Properties.ReasonString)
		return nil, nil
	}
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		logger.Fatalf("failed to create client, %v", err)
		return nil, err
	}
	mc, err := &MQTTConnection {
		client:		client,
		logger:		logger,
		addr:		addr,
		topic:		topic,
		ceClient:	c,
	}

	return mc, nil
}

func (mc *MQTTConnection) Subscribe(topic string) {
	_, err = c.Subscribe(context.Background(), &paho.Subscribe{
			Subscriptions: map[string]paho.SubscribeOptions{
				topic: paho.SubscribeOptions{QoS: 0},
			},
	})
	mc.logger.Infof("Subscribed to topic %s", topic)
}

func (mc *MQTTConnection) OnReceive(client mqtt.Client, msg mqtt.Message) {
	event := cloudevents.NewEvent()
	event.SetData(cloudevents.ApplicationJSON, msg)

}

type ConnectionManager struct {
	conn				map[types.NamespacedName]*MQTTConnection
	logger				*zap.SugaredLogger
}

func (cm *ConnectionManager) AddConn(obj interface{}) {
	bc := obj.(*v1alpha1.BrokerChannel)
	ID := types.NamespacedName{Namespace: bc.Namespace, Name: bc.Name}
	mc, ok := cm.conn[ID]
	if !ok {
		newConn, err := newMQTTConnection(bc.Status.Address.URL, bc.Spec.Topic, bc.Spec.BrokerAddr, bc.Spec.BrokerPort, cm.logger)
		if err != nil {
			panic(err)
		}
		cm.conn[ID] = newConn
	}
	mc.client.Subscribe(bc.Spec.Topic, 0, mc.OnReceive)
}

func (cm *ConnectionManager) DeleteConn(obj interface{}) {
	bc := obj.(*v1alpha1.BrokerChannel)
	ID := types.NamespacedName{Namespace: bc.Namespace, Name: bc.Name}
	delete(cm.conn, ID)
}


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up our logger.
	loggingConfig, err := sharedmain.GetLoggingConfig(ctx)
	if err != nil {
		log.Fatal("Error loading/parsing logging configuration: ", err)
	}

	logger, _ := logging.NewLoggerFromConfig(loggingConfig, "MQTTCoverter")
	logger = logger.With(zap.String(logkey.ControllerType, "MQTTConverter"))
	ctx = logging.WithLogger(ctx, logger)

	cm := &ConnectionManager {
		conn: make(map[types.NamespacedName]*MQTTConnection),
		logger: logger,
	}

	brokerChannelInformer := brokerchannelinformer.Get(ctx)
	brokerChannelInformer.Informer().AddEventHandle(cache.ResourceEventHandlerFuncs{
		AddFunc: cm.AddConn,
		UpdateFunc: controller.PassNew(cm.AddConn),
		DeleteFunc: cm.DeleteConn,
	})
	sigCh := signals.SetupSignalHandler()
	select {
	case <-sigCh:
		logger.Info("Received SIGTERM")
		// Send a signal to let readiness probes start failing.
		cancel()
	}
	logger.Info("terminated")
}
