package gin

import (
	"github.com/colonyos/colonies/pkg/backends"
	ginframework "github.com/gin-gonic/gin"
)

// GinBackendFactory implements the backends.BackendFactory interface by
// delegating to the existing package-level functions in this package.
type GinBackendFactory struct{}

// Compile-time check that GinBackendFactory satisfies BackendFactory.
var _ backends.BackendFactory = (*GinBackendFactory)(nil)

func (f *GinBackendFactory) CreateEngine() backends.Engine {
	return CreateEngine()
}

func (f *GinBackendFactory) CreateServer(port int, engine backends.Engine) backends.Server {
	return NewBackendServer(port, engine)
}

func (f *GinBackendFactory) CreateRealtimeHandler() backends.RealtimeEventHandler {
	return CreateEventHandler(nil)
}

func (f *GinBackendFactory) CreateTestableRealtimeHandler() backends.TestableRealtimeEventHandler {
	return CreateTestableEventHandler(nil)
}

func (f *GinBackendFactory) ConfigureSilentMode() {
	ginframework.SetMode(ginframework.ReleaseMode)
}

func (f *GinBackendFactory) CORS() backends.HandlerFunc {
	middleware := CORS()
	return func(c backends.Context) {
		middleware(c)
	}
}

func init() {
	backends.RegisterBackend("gin", &GinBackendFactory{})
}
