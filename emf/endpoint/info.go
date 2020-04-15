package endpoint

import (
	"net/http"
	"time"

	"github.com/cambridge-blockchain/emf/models"
)

// Build represents the versioning and time stamp for the binary run by the component.
type Build struct {
	Version          string `json:"version"`
	ReleaseTimestamp string `json:"release_timestamp"`
	Build            string `json:"build"`
	EMFVersion       string `json:"emf_version"`
	EchoVersion      string `json:"echo_version"`
}

// Caller contains relevant information about a request.
type Caller struct {
	RequestID  string `json:"request_id"`
	RemoteAddr string `json:"remote_addr"`
	RequestURI string `json:"request_uri"`
	Referer    string `json:"referer"`
	UserAgent  string `json:"user_agent"`
}

// Information encapusaltes relevant information about the running component and the client request.
type Information struct {
	CurrentTime           time.Time          `json:"current_time"`
	ComponentName         string             `json:"component_name"`
	APIPort               string             `json:"api_port"`
	Build                 Build              `json:"build"`
	Caller                Caller             `json:"caller"`
	AdditionalInformation []models.DataPoint `json:"additional_information"`
}

// RegisterInfo registers the identity endpoints to the provided group
func RegisterInfo(r *models.Router, bc models.BuildConfig, mids ...models.Middleware) {
	var g = r.NewGroup("/info", mids...)
	g.GET("", wrapGetInfo(bc))
}

func wrapGetInfo(bc models.BuildConfig) models.HandlerFunc {
	return func(c models.Context) (err error) {
		var (
			info Information
		)

		info.CurrentTime = time.Now().UTC()
		info.ComponentName = bc.Component
		info.APIPort = c.Echo().Server.Addr
		info.Caller = getCallerInfo(c)
		info.Build = getBuildInfo(bc)

		if bc.AdditionalInformation != nil {
			info.AdditionalInformation = bc.AdditionalInformation
		} else {
			info.AdditionalInformation = []models.DataPoint{}
		}

		return c.JSON(http.StatusOK, info)
	}
}

func getCallerInfo(c models.Context) Caller {
	var (
		req        *http.Request
		remoteAddr string
	)

	req = c.Request()
	if remoteAddr = req.Header.Get("x-real-ip"); remoteAddr == "" {
		remoteAddr = req.RemoteAddr
	}

	return Caller{
		RequestID:  c.GetRequestID(),
		RemoteAddr: remoteAddr,
		RequestURI: req.RequestURI,
		Referer:    req.Referer(),
		UserAgent:  req.UserAgent(),
	}
}

func getBuildInfo(bc models.BuildConfig) Build {
	return Build{
		Version:          bc.Version,
		EMFVersion:       bc.EMFVersion,
		EchoVersion:      bc.EchoVersion,
		ReleaseTimestamp: bc.ReleaseTimestamp,
		Build:            bc.Build,
	}
}
