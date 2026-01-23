package items

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// ItemFromJSONString ...
func ItemFromJSONString(input string) (*Item, error) {
	bytes := []byte(input)
	var item Item
	if err := json.Unmarshal(bytes, &item); err != nil {
		logrus.WithField("input", input).WithError(err).Info("Could not unmarshal item from string")
		return nil, err
	}
	return &item, nil
}

// ItemsFromJSONString ...
func ItemsFromJSONString(input string) (Items, error) {
	bytes := []byte(input)
	var its Items
	if err := json.Unmarshal(bytes, its); err != nil {
		logrus.WithField("input", input).WithError(err).Info("Could not unmarshal items array from string")
		return nil, err
	}
	return its, nil
}

// ItemToJSONString ...
func ItemToJSONString(item Item) string {
	bytes, _ := json.MarshalIndent(item, "", " ")
	return string(bytes)
}

// ItemsToJSONString ...
func ItemsToJSONString(items Items) string {
	bytes, _ := json.MarshalIndent(items, "", " ")
	return string(bytes)
}
