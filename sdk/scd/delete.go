package scd

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

func (s *Service) DeleteOperationalIntent(ctx context.Context, id scdussv1.EntityID) error {
	// TODO(gap): No validation; errors are ignored
	intent, _ := s.Intents.Get(ctx, id)
	s.DSS.DeleteOperationalIntentReference(ctx, intent.EntityID, intent.OVN)
	return s.Intents.Delete(ctx, id)
}
