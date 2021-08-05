package main

import (
	"net"
	"context"
	"fmt"
	"log"
	"sync"

	"k8s.io/client-go/tools/cache"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/logging/logkey"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/injection"

	v1alpha1 "github.com/ShixiongQi/brokerchannel/pkg/apis/samples/v1alpha1"
	brokerchannelinformer "github.com/ShixiongQi/brokerchannel/pkg/client/injection/informers/samples/v1alpha1/brokerchannel"
	"knative.dev/pkg/apis"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type MQTTConnection struct {
	client *paho.Client
	logger *zap.SugaredLogger
	topic string
	addr *apis.URL
	ceClient	cloudevents.Client
	stopCh		<-chan struct{}
}

func newMQTTConnection(addr *apis.URL, topic, broker string, port int, logger *zap.SugaredLogger, stopCh <-chan struct{}) (*MQTTConnection, error) {
	server := fmt.Sprintf("%s:%d", broker, port)
	logger.Infof("Create connection to %s", server)
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		logger.Fatalf("failed to create client, %v", err)
		return nil, err
	}
	mc := &MQTTConnection {
		client:		paho.NewClient(),
		logger:		logger,
		addr:		addr,
		topic:		topic,
		ceClient:	c,
		stopCh:		stopCh,
	}
	mc.client.Conn = conn
	logger.Infof("Url is %v", addr)
	mc.client.Router = paho.NewSingleHandlerRouter(func(m *paho.Publish) {
			// logger.Infof("Receive object %v\n", m)
			event := cloudevents.NewEvent()
			prop := m.Properties.User
			event.SetSource(prop["source"])
			event.SetType(prop["type"])
			event.SetID(prop["ID"])
			// logger.Infof("QLOG: source: %v, type: %v, ID: %v", prop["source"], prop["type"], prop["ID"])
			event.SetData(cloudevents.ApplicationJSON, m.Payload)
			ctx := cloudevents.ContextWithTarget(context.Background(), mc.addr.URL().String())
			if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
				mc.logger.Fatalf("failed to send, %v", result)
			}
		})
	ca, err := mc.client.Connect(context.Background(), &paho.Connect{CleanStart: true, KeepAlive:  30,})
	if err != nil {
		return nil, err
	}
	if ca.ReasonCode != 0 {
		logger.Errorf("Failed to connect to %s : %d - %s", server, ca.ReasonCode, ca.Properties.ReasonString)
		return nil, nil
	}

	return mc, nil
}

func (mc *MQTTConnection) Run(wg *sync.WaitGroup) {
	mc.logger.Info("Start routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-mc.stopCh
		mc.client.Disconnect(&paho.Disconnect{ReasonCode: 0})
		mc.logger.Info("Disconnected")
	}()
}

func (mc *MQTTConnection) Subscribe(ctx context.Context, topic string) error {
	if _, err := mc.client.Subscribe(ctx, &paho.Subscribe{Subscriptions: map[string]paho.SubscribeOptions{
		topic: {QoS: 0},
	},}); err != nil {
		return err
	}
	return nil
}
type ConnectionManager struct {
	conn				map[types.NamespacedName]*MQTTConnection
	logger				*zap.SugaredLogger
	sigCh				<-chan struct{}
	wg					*sync.WaitGroup
	ctx					context.Context
}

func (cm *ConnectionManager) AddConn(obj interface{}) {
	bc := obj.(*v1alpha1.BrokerChannel)
	ID := types.NamespacedName{Namespace: bc.Namespace, Name: bc.Name}
	_, ok := cm.conn[ID]
	if !ok {
		newConn, err := newMQTTConnection(bc.Status.SinkURI, bc.Spec.Topic, bc.Spec.BrokerAddr, bc.Spec.BrokerPort, cm.logger, cm.sigCh)
		if err != nil {
			panic(err)
		}
		cm.conn[ID] = newConn
	}
	cm.conn[ID].addr = bc.Status.SinkURI
	if err := cm.conn[ID].Subscribe(cm.ctx, bc.Spec.Topic); err != nil {
		panic(err)
	}
	cm.conn[ID].Run(cm.wg)
}

func (cm *ConnectionManager) DeleteConn(obj interface{}) {
	cm.logger.Info("Connection stopped")
	bc := obj.(*v1alpha1.BrokerChannel)
	ID := types.NamespacedName{Namespace: bc.Namespace, Name: bc.Name}
	delete(cm.conn, ID)
}


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer cancel()

	metrics.MemStatsOrDie(ctx)

	cfg := injection.ParseAndGetRESTConfigOrDie()
	ctx, informers := injection.Default.SetupInformers(ctx, cfg)
	// Set up our logger.
	loggingConfig, err := sharedmain.GetLoggingConfig(ctx)
	if err != nil {
		log.Fatal("Error loading/parsing logging configuration: ", err)
	}

	logger, _ := logging.NewLoggerFromConfig(loggingConfig, "MQTTCoverter")
	logger = logger.With(zap.String(logkey.ControllerType, "MQTTConverter"))
	ctx = logging.WithLogger(ctx, logger)
	if err := controller.StartInformers(ctx.Done(), informers...); err != nil {
		logger.Fatalw("Failed to start informers", zap.Error(err))
	}
	sigCh := signals.SetupSignalHandler()

	cm := &ConnectionManager {
		conn: make(map[types.NamespacedName]*MQTTConnection),
		logger: logger,
		sigCh: sigCh,
		wg:		&wg,
		ctx:	ctx,
	}

	brokerChannelInformer := brokerchannelinformer.Get(ctx)
	brokerChannelInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cm.AddConn,
		UpdateFunc: controller.PassNew(cm.AddConn),
		DeleteFunc: cm.DeleteConn,
	})
	select {
	case <-sigCh:
		logger.Info("Received SIGTERM")
		// Send a signal to let readiness probes start failing.
		cancel()
	}
	wg.Wait()
	logger.Info("terminated")
}
