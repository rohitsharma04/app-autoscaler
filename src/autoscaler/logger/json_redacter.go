package logger

import (
	"encoding/json"
	"regexp"
	"strings"

	"code.cloudfoundry.org/lager"
)

const postgresDbURLPattern = "^(postgresq?l?:\\/\\/)([\\da-zA-Z]+):([\\da-zA-Z]+)@([\\da-zA-Z\\.-]+)(:[\\d]{4,5})?\\/([\\da-zA-Z]+)"

type JSONRedacterWithURLCred struct {
	jsonRedacter   *lager.JSONRedacter
	urlCredMatcher *regexp.Regexp
}

func NewJSONRedacterWithURLCred(keyPatterns []string, valuePatterns []string) (*JSONRedacterWithURLCred, error) {
	jsonRedacter, err := lager.NewJSONRedacter(keyPatterns, valuePatterns)
	if err != nil {
		return nil, err
	}
	urlCredMatcher, err := regexp.Compile(postgresDbURLPattern)
	return &JSONRedacterWithURLCred{
		jsonRedacter:   jsonRedacter,
		urlCredMatcher: urlCredMatcher,
	}, nil
}

func (r JSONRedacterWithURLCred) Redact(data []byte) []byte {
	var jsonBlob interface{}
	err := json.Unmarshal(data, &jsonBlob)
	if err != nil {
		return handleError(err)
	}
	r.redactValue(&jsonBlob)

	data, err = json.Marshal(jsonBlob)
	if err != nil {
		return handleError(err)
	}

	return r.jsonRedacter.Redact(data)
}

func (r JSONRedacterWithURLCred) redactURL(url string) string {
	userNameStart := strings.IndexAny(url, ":") + 2
	passwordStart := userNameStart + strings.IndexAny(url[userNameStart:], ":")
	passwordEnd := passwordStart + strings.IndexAny(url[passwordStart:], "@")
	return url[:passwordStart+1] + "*REDACTED*" + url[passwordEnd:]
}

func (r JSONRedacterWithURLCred) redactValue(data *interface{}) interface{} {
	if data == nil {
		return data
	}

	if a, ok := (*data).([]interface{}); ok {
		r.redactArray(&a)
	} else if m, ok := (*data).(map[string]interface{}); ok {
		r.redactObject(&m)
	} else if s, ok := (*data).(string); ok {
		if r.urlCredMatcher.MatchString(s) {
			(*data) = r.redactURL(s)
		}
	}
	return (*data)
}

func (r JSONRedacterWithURLCred) redactArray(data *[]interface{}) {
	for i, _ := range *data {
		r.redactValue(&((*data)[i]))
	}
}

func (r JSONRedacterWithURLCred) redactObject(data *map[string]interface{}) {
	for k, v := range *data {
		(*data)[k] = r.redactValue(&v)
	}
}

func handleError(err error) []byte {
	var content []byte
	if _, ok := err.(*json.UnsupportedTypeError); ok {
		data := map[string]interface{}{"lager serialisation error": err.Error()}
		content, err = json.Marshal(data)
	}
	if err != nil {
		panic(err)
	}
	return content
}
