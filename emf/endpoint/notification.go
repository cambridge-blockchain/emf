package endpoint

import (
	"net/http"

	"github.com/cambridge-blockchain/emf/notifications"

	"github.com/cambridge-blockchain/emf/models"
)

// RegisterNotification registers the endpoint required by notification-component to pull
//	notification codes from all other components
func RegisterNotification(r *models.Router, notificationCodes []notifications.NotificationType,
	mids ...models.Middleware) {
	var g = r.NewGroup("/noauth/notifications/list", mids...)
	g.GET("", ListNotifications(notificationCodes))
}

// ListNotifications simply returns the set of notification codes used by the component
func ListNotifications(notificationCodes []notifications.NotificationType) models.HandlerFunc {
	return func(c models.Context) (err error) {
		type notification struct {
			Code []string `json:"codes"`
		}

		var noti notification
		for _, n := range notificationCodes {
			noti.Code = append(noti.Code, n.Code)
		}
		return c.JSON(http.StatusOK, noti)
	}
}
