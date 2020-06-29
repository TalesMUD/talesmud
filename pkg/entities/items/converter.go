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


// ItemTemplateFromJSONString ...
func ItemTemplateFromJSONString(input string) (*ItemTemplate, error) {
	bytes := []byte(input)
	var itemTemplate ItemTemplate
	if err := json.Unmarshal(bytes, &itemTemplate); err != nil {
		logrus.WithField("input", input).WithError(err).Info("Could not unmarshal itemtemplate from string")
		return nil, err
	}
	return &itemTemplate, nil
}

// ItemTemplatesFromJSONString ...
func ItemTemplatesFromJSONString(input string) (ItemTemplates, error) {
	bytes := []byte(input)
	var its ItemTemplates
	if err := json.Unmarshal(bytes, its); err != nil {
		logrus.WithField("input", input).WithError(err).Info("Could not unmarshal itemtemplate array from string")
		return nil, err
	}
	return its, nil
}

// ItemTemplateToJSONString ...
func ItemTemplateToJSONString(itemTemplate ItemTemplate) string {
	bytes, _ := json.MarshalIndent(itemTemplate, "", " ")
	return string(bytes)
}

// ItemTemplatesToJSONString ...
func ItemTemplatesToJSONString(itemTemplates ItemTemplates) string {
	bytes, _ := json.MarshalIndent(itemTemplates, "", " ")
	return string(bytes)
}
