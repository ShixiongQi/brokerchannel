# Broker Channel - An MQTT to HTTP Adaptor

Inspired by Knative SampleSource, an extension to MQTT broker
## How tro install
1. Install [ko](https://github.com/google/ko.git)
2. Install namespaces, deployments, service under `config` using `ko`
	* Except `brokerchannel-service.yaml`, the naming is smiliar to Knative repo, install the yaml with numbered filenames.
	* `ko apply -f <filename>`
	* Webhook is not yet implemented, you can skip files whose name contains webhook 
3. For CRD, please apply `brokerchannel1.yaml` instead of `brokerchannel.yaml`
4. Go back to project root, checkout `example.yaml` and `temp.yaml` to see how to setup a subscription
