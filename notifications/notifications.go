package notifications

import (
	"fmt"
	"net/http"

	"github.com/cambridge-blockchain/emf/models"
)

//NotificationMDField is used to represented a Metadata parameter used for templating a message
type NotificationMDField struct {
	Name  string
	value string
}

//Is sets the 'Value' field of the notificationMDField
func (n NotificationMDField) Is(value string) NotificationMDField {
	return NotificationMDField{
		Name:  n.Name,
		value: value,
	}
}

//TemplateName returns the template version of the MDField's name
func (n NotificationMDField) TemplateName() string {
	return fmt.Sprintf("{{.%s}}", n.Name)
}

//NotificationType is a struct to represent a non-instantiated notification
type NotificationType struct {
	Code string
	//IsEnabled bool
	Fields   []NotificationMDField
	Template string
}

//NotificationPayload is a struct to represent an instantiated notification of a specific type
type NotificationPayload struct {
	ToEntity   string            `json:"toEntity"`
	FromEntity string            `json:"fromEntity"`
	Template   string            `json:"template"`
	Metadata   map[string]string `json:"metadata"`
	Code       string            `json:"code"`
}

//Send creates a new instantiation of the given notification type with the supplied metadata and sends it
func (n NotificationType) Send(rh models.RequestHandler, to, from string, mdFields ...NotificationMDField) {
	go func(rh models.RequestHandler, payload NotificationPayload) {
		if err := sendNotification(rh, payload); err != nil {
			rh.Logger().Debugf("Error sending notification payload %+v\nErr: %s", n, err.Error())
		}
	}(rh, NotificationPayload{
		ToEntity:   to,
		FromEntity: from,
		Template:   n.Template,
		Metadata:   notificationFieldsToMap(mdFields),
		Code:       n.Code,
	})
}

//notificationFieldsToMap is responsible for converting the list of NotificationMDFields into a map[string]string
func notificationFieldsToMap(params []NotificationMDField) (m map[string]string) {
	m = make(map[string]string)
	for _, k := range params {
		m[k.Name] = k.value
	}
	return m
}

// sendNotification sends a notification to the notification component.
func sendNotification(rh models.RequestHandler, nPayload NotificationPayload) (err error) {
	if err = rh.Requester(
		http.MethodPost,
		"notification",
		"/notifications",
		nPayload,
		&map[string]interface{}{},
	); err != nil {
		return fmt.Errorf("failed to push notification to graph. Err: %+v", err)
	}

	return
}
