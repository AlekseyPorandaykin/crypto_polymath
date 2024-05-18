package ui

import "context"

type UI interface {
	Run(ctx context.Context) error
}
