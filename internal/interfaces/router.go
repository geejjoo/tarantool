package interfaces

import "context"

type Router interface {
	Run(addr string) error
	Shutdown(ctx context.Context) error
}
