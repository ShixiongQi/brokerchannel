# Broker Channel - An MQTT to HTTP Adaptor

Inspired by Knative SampleSource, an extension to MQTT broker
## How to install
1. Install [ko](https://github.com/google/ko.git)
2. Install namespaces, deployments, service under `config` using `ko`
	* Webhook is not yet implemented, you can skip files whose name contains webhook
	```
	ko apply -f 100-namespace.yaml
	ko apply -f 200-serviceaccount.yaml
	ko apply -f 201-clusterrole.yaml
	ko apply -f 202-clusterrolebinding.yaml
	ko apply -f 400-controller-service.yaml
	ko apply -f 500-controller.yaml
	ko apply -f brokerchannel-crd.yaml # Install CRD definition
	ko apply -f brokerchannel-service.yaml # Install MQTT-to-HTTP adaptor
	```
3. Go back to project root, checkout `temp.yaml` to see how to setup a subscription
