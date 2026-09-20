package scd

import (
	"context"
	"errors"

	"bwawan.com/openuss/internal/api/scdussv1"
)

func (s *Service) DeleteOperationalIntent(ctx context.Context, id scdussv1.EntityID) error {
	intent, err := s.Intents.Get(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	// TODO: What if there is a different kind of error?
	_, err = s.DSS.DeleteOperationalIntentReference(ctx, intent.EntityID, intent.OVN)
	if err != nil {
		return err
	}
	// TODO: Missing context
	return s.Intents.Delete(nil, id)
}
