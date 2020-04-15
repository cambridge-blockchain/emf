// Package cache provides wrappers for better interacting with the underlying cache implementation
package cache

import (
	"encoding/json"
	"net/http"
	"strings"

	emf "github.com/cambridge-blockchain/emf/models"
)

// Client is the interface for any application level cache implementation.
type Client interface {
	Set(path string, data []byte) error
	SetEx(path string, data []byte) error
	Get(path string) ([]byte, bool, error)
	Delete(path string) error
}

// CacheableRequestHandler is a request handler returning a CacheResponse
type CacheableRequestHandler func(emf.Context) (*CacheResponse, error)

// PreRequestFunc can be used before the request handler is called
type PreRequestFunc func(c emf.Context, ctx Client, ce CacheableEntity) ([]byte, error)

// PostRequestFunc can be used after the reequest handler is called
type PostRequestFunc func(c emf.Context, ctx Client, ce CacheableEntity, cr CacheResponse) error

// CacheableEntity interface must be impemented for response objects that use the CacheWrapper.
type CacheableEntity interface {
	GetCacheKey(c emf.Context) string
	GetCacheKeysFromEntityID(entityID string) []string
}

// DefaultCacheableEntity can be embedded into a struct so that it fulfills the CacheableEntity
// interface using the  default behavior.
type DefaultCacheableEntity struct{}

// CacheResponse is the response needed by handlers using the CacheWrapper.
type CacheResponse struct {
	code             int
	response         interface{}
	invalidEntityIDs []string
}

// AddIDToInvalidate adds the id to the list of ids to be invalidated.
func (c *CacheResponse) AddIDToInvalidate(id string) {
	c.invalidEntityIDs = append(c.invalidEntityIDs, id)
}

// SetResponseCode sets the http response code
func (c *CacheResponse) SetResponseCode(code int) {
	c.code = code
}

// GetResponseCode gets the http response code
func (c CacheResponse) GetResponseCode() int {
	return c.code
}

// SetResponse sets response to supplied interface
func (c *CacheResponse) SetResponse(r interface{}) {
	c.response = r
}

// GetResponse gets the response
func (c CacheResponse) GetResponse() interface{} {
	return c.response
}

// NewCacheResponse returns a new CacheResponse
func NewCacheResponse(r interface{}, code int) *CacheResponse {
	return &CacheResponse{
		code:     code,
		response: r,
	}
}

// NewCacheResponseOK returns a new CacheResponse
func NewCacheResponseOK(r interface{}) *CacheResponse {
	return &CacheResponse{
		code:     http.StatusOK,
		response: r,
	}
}

// NewEmptyCacheResponse returns a new CacheResponse
func NewEmptyCacheResponse(code int) *CacheResponse {
	return &CacheResponse{
		code: code,
	}
}

// GetCacheKey satisfies the default CacheableEntity.
func (d *DefaultCacheableEntity) GetCacheKey(c emf.Context) string {
	return getDefaultCacheKey(c)
}

// getDefaultCacheKey returns the full GET & PUT URL path with params.
func getDefaultCacheKey(c emf.Context) string {
	return c.Request().URL.String()
}

// GetCacheKeysFromEntityID satisfies the default CacheableEntity.
func (d *DefaultCacheableEntity) GetCacheKeysFromEntityID(id string) []string {
	return []string{id}
}

// GetBaseCacheKey returns the base routed URL to be used for a more general cache key.
func GetBaseCacheKey(c emf.Context) string {
	var cachepath = c.Path()
	for i, p := range c.ParamNames() {
		cachepath = strings.Replace(cachepath, ":"+p, c.ParamValues()[i], -1)
	}
	return cachepath
}

// GetRoleBasedCacheKey returns the cache path for URL with role
// May result in prefixing the cache key with pds, tp, sp, or pa whom may have different
// views of the requested entity.
func GetRoleBasedCacheKey(c emf.Context) string {
	role, _ := c.GetClaim("rol") // nolint:errcheck
	return strings.ToLower(strings.Split(role, "-")[0]) + getDefaultCacheKey(c)
}

// DefaultCacheHandlerWrapper handles caching and cache-breaking intended to work across
// GET, PUT & POST action endpoints we support.
func DefaultCacheHandlerWrapper(
	cacheClient Client,
	cacheableEntity CacheableEntity,
	requestHandlerFunc CacheableRequestHandler,
) func(c emf.Context) (err error) {
	return CacheHandlerWrapper(cacheClient, cacheableEntity, requestHandlerFunc,
		DefaultPreRequestFunc, DefaultPostRequestFunc)
}

// CacheHandlerWrapper handles wrapping typical echo request handlers to provide request based
// caching and cache-breaking.
// The CachebleEntity provides the functionality for generating the cache key based on the request.
// The CacheableRequestHandler is a request handler formatted to respond in to the cache handler wrapper.
// preRequestFunc and postRequestFunc are called before and after the request respectively when provided.
func CacheHandlerWrapper(
	cacheClient Client,
	cacheableEntity CacheableEntity,
	requestHandlerFunc CacheableRequestHandler,
	preRequestFunc PreRequestFunc,
	postRequestFunc PostRequestFunc,
) func(c emf.Context) error {
	return func(c emf.Context) error {
		if preRequestFunc != nil && cacheClient != nil {
			if cachedValue, err := preRequestFunc(c, cacheClient, cacheableEntity); err != nil {
				c.Logger().Errorf("cache error in pre-request processing: %+v", err)
			} else if cachedValue != nil {
				return c.JSONBlob(http.StatusOK, cachedValue)
			}
		}

		cr, err := requestHandlerFunc(c)
		if err != nil {
			return err
		}

		if postRequestFunc != nil && cacheClient != nil {
			if err := postRequestFunc(c, cacheClient, cacheableEntity, *cr); err != nil {
				c.Logger().Errorf("cache error in post-request processing: %+v", err)
			}
		}

		return c.JSON(cr.code, cr.response)
	}
}

// DefaultPreRequestFunc requests entity from the cached based on the CacheableEntity
func DefaultPreRequestFunc(c emf.Context, ctx Client, ce CacheableEntity) ([]byte, error) {
	key := ce.GetCacheKey(c)
	if c.Request().Method == http.MethodGet {
		if b, exists, err := ctx.Get(key); err != nil {
			c.Logger().Errorf("error getting %s from cache: %+v", key, err)
		} else if exists && b != nil {
			return b, nil
		}
	}

	return nil, nil
}

// DefaultPostRequestFunc invalidates entries in the cache based on the CacheableEntity
func DefaultPostRequestFunc(c emf.Context, ctx Client, ce CacheableEntity, cr CacheResponse) (err error) {
	key := ce.GetCacheKey(c)

	switch c.Request().Method {
	case http.MethodGet:
		if cr.code == http.StatusOK {
			var b []byte
			if b, err = json.Marshal(cr.response); err != nil {
				c.Logger().Errorf("error marshaling response: \n%+v %+v", cr.response, err)
				return
			}
			if err = ctx.Set(key, b); err != nil {
				c.Logger().Errorf("error setting %s to cache: %+v", key, err)
				return
			}
		}
	case http.MethodPut:
		if err = ctx.Delete(key); err != nil {
			c.Logger().Errorf("error deleting %s from cache: %+v", key, err)
			return
		}
	case http.MethodPost:
		if strings.HasSuffix(c.Path(), "/action") {
			// POST /action: invalidate potentially multiple entries likely depending on request response
			if len(cr.invalidEntityIDs) == 0 {
				c.Logger().Warnf("warning no cache values removed after POST %s", c.Request().URL.String())
			}
			for _, id := range cr.invalidEntityIDs {
				// todo: delete multiple keys at same time
				for _, key := range ce.GetCacheKeysFromEntityID(id) {
					if err := ctx.Delete(key); err != nil {
						c.Logger().Errorf("error deleting %s from cache: %+v", key, err)
					}
				}
			}
		}
	}

	return nil
}
