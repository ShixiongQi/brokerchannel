module github.com/ShixiongQi/brokerchannel

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/eclipse/paho.golang v0.9.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.16.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	knative.dev/eventing v0.22.1-0.20210420082335-31cda4cec54b
	knative.dev/hack v0.0.0-20210325223819-b6ab329907d3
	knative.dev/pkg v0.0.0-20210420053235-1afd04993622
)

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
