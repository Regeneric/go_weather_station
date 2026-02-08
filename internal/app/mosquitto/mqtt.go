package mosquitto

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/Regeneric/go_weather_station/internal/config"
	"github.com/Regeneric/go_weather_station/internal/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

func MQTTPublish(data <-chan utils.SystemSnapshot, client mqtt.Client) {
	log := slog.With("func", "MQTTSender()", "params", "(<-chan SystemSnapshot, mqtt.Client)", "package", "main", "module", "mqtt")
	log.Debug("Publishing data to MQTT topic")

	if config.MQTTEnable == false {
		log.Warn("MQTT has been disabled in the config file!", "enable", config.MQTTEnable)
		return
	}

	for snapshot := range data {
		slog.Debug("Processing system snapshot", "items", len(snapshot))

		for sensorID, payloadData := range snapshot {
			jsonPayload, err := json.Marshal(payloadData)
			if err != nil {
				continue
			}

			topic := fmt.Sprintf("%s/%s", config.MQTTopic, sensorID)
			client.Publish(topic, 0, false, jsonPayload)

			slog.Debug("Data published to MQTT", "topic", topic)
		}
		slog.Debug("MQTT sent update", "sensors", len(snapshot))
	}
}

func MQTTInit() mqtt.Client {
	log := slog.With("func", "Init()", "params", "(-)", "package", "main", "module", "mqtt")
	log.Debug("Initilizing MQTT client")

	if config.MQTTEnable == false {
		log.Warn("MQTT has been disabled in the config file!", "enable", config.MQTTEnable)
		return nil
	}

	_ = godotenv.Load()
	opts := mqtt.NewClientOptions()

	connectionString := fmt.Sprintf("tcp://%s:%s", config.MQTTBrokerAddress, config.MQTTBrokerPort)
	opts.AddBroker(connectionString)

	opts.SetClientID(config.MQTTDeviceName)
	opts.SetUsername(os.Getenv("MQTT_USERNAME"))
	opts.SetPassword(os.Getenv("MQTT_PASSWORD"))

	opts.SetAutoReconnect(config.MQTTAutoReconnect)
	opts.SetMaxReconnectInterval(config.MQTTReconnectInterval)
	opts.SetKeepAlive(config.MQTTKeepAlive)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.Info("MQTT connected", "server", connectionString, "device", config.MQTTDeviceName)
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		slog.Warn("MQTT connection lost!", "server", connectionString, "device", config.MQTTDeviceName)
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return client
}
