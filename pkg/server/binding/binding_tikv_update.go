package binding

import (
	"context"
	"errors"

	"github.com/opensibyl/sibyl2/pkg/server/object"
)

func (t *tikvDriver) UpdateRevProperties(wc *object.WorkspaceConfig, k string, v any, ctx context.Context) error {
	// TODO implement me
	return errors.New("implement me")
}

func (t *tikvDriver) UpdateFileProperties(wc *object.WorkspaceConfig, path string, k string, v any, ctx context.Context) error {
	// TODO implement me
	return errors.New("implement me")
}

func (t *tikvDriver) UpdateFuncProperties(wc *object.WorkspaceConfig, signature string, k string, v any, ctx context.Context) error {
	// TODO implement me
	return errors.New("implement me")
}
