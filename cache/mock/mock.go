package mock

import (
	"github.com/cambridge-blockchain/emf/cache"
	emf "github.com/cambridge-blockchain/emf/models"
)

// RequestWrapper exposes a CacheableRequestHandler like a normal echo  request handler.
func RequestWrapper(c emf.Context, f cache.CacheableRequestHandler) error {
	cr, err := f(c)
	if err != nil {
		return err
	}

	return c.JSON(cr.GetResponseCode(), cr.GetResponse())
}
