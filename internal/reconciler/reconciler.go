// Package reconciler implements the asset matching and change detection engine.
// When a collector brings in new data, the reconciler matches incoming assets to
// existing ones, detects attribute changes, and publishes events.
package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/model"
	"github.com/qrunner/arch/internal/store/postgres"
	neostore "github.com/qrunner/arch/internal/store/neo4j"
	"go.uber.org/zap"
)

// EventPublisher defines the interface for publishing change events.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// Reconciler matches incoming assets against existing records and detects changes.
type Reconciler struct {
	pgStore   *postgres.Store
	neoStore  *neostore.Store
	publisher EventPublisher
	logger    *zap.Logger
}

// New creates a new Reconciler.
func New(pgStore *postgres.Store, neoStore *neostore.Store, publisher EventPublisher, logger *zap.Logger) *Reconciler {
	return &Reconciler{
		pgStore:   pgStore,
		neoStore:  neoStore,
		publisher: publisher,
		logger:    logger,
	}
}

// Reconcile processes a CollectResult, matching and upserting assets and relationships.
func (r *Reconciler) Reconcile(ctx context.Context, result *collector.CollectResult) error {
	for i := range result.Assets {
		incoming := &result.Assets[i]
		if err := r.reconcileAsset(ctx, incoming); err != nil {
			r.logger.Error("reconciling asset",
				zap.String("external_id", incoming.ExternalID),
				zap.String("source", incoming.Source),
				zap.Error(err),
			)
			continue
		}
	}

	for i := range result.Relationships {
		rel := &result.Relationships[i]
		if err := r.reconcileRelationship(ctx, rel); err != nil {
			r.logger.Error("reconciling relationship",
				zap.String("from", rel.FromID.String()),
				zap.String("to", rel.ToID.String()),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

func (r *Reconciler) reconcileAsset(ctx context.Context, incoming *model.Asset) error {
	// Step 1: Try exact match by external_id + source
	existing, err := r.pgStore.GetByExternalID(ctx, incoming.Source, incoming.ExternalID)
	if err != nil {
		return fmt.Errorf("looking up asset: %w", err)
	}

	// Step 2: If no exact match, try fuzzy match by IP
	if existing == nil && len(incoming.IPAddresses) > 0 {
		for _, ip := range incoming.IPAddresses {
			matches, err := r.pgStore.FindByIP(ctx, ip)
			if err != nil {
				r.logger.Warn("fuzzy IP match failed", zap.Error(err))
				continue
			}
			if len(matches) == 1 {
				existing = matches[0]
				break
			}
		}
	}

	// Step 3: If no match by IP, try fuzzy match by FQDN
	if existing == nil && incoming.FQDN != nil && *incoming.FQDN != "" {
		matches, err := r.pgStore.FindByFQDN(ctx, *incoming.FQDN)
		if err != nil {
			r.logger.Warn("fuzzy FQDN match failed", zap.Error(err))
		} else if len(matches) == 1 {
			existing = matches[0]
		}
	}

	now := time.Now().UTC()

	if existing == nil {
		// New asset
		incoming.FirstSeen = now
		incoming.LastSeen = now
		incoming.Status = model.AssetStatusActive
		incoming.CreatedAt = now
		incoming.UpdatedAt = now

		if err := r.pgStore.Create(ctx, incoming); err != nil {
			return fmt.Errorf("creating new asset: %w", err)
		}

		if r.neoStore != nil {
			if err := r.neoStore.UpsertNode(ctx, incoming); err != nil {
				r.logger.Warn("failed to create neo4j node", zap.Error(err))
			}
		}

		r.publishEvent(ctx, model.ChangeActionCreated, incoming, nil)
		return nil
	}

	// Existing asset - detect changes
	diff := detectChanges(existing, incoming)
	existing.LastSeen = now
	existing.Status = model.AssetStatusActive

	if len(diff) > 0 {
		// Apply changes
		existing.Name = incoming.Name
		existing.FQDN = incoming.FQDN
		existing.IPAddresses = incoming.IPAddresses
		existing.Attributes = incoming.Attributes
		existing.UpdatedAt = now

		if err := r.pgStore.Update(ctx, existing); err != nil {
			return fmt.Errorf("updating asset: %w", err)
		}

		if r.neoStore != nil {
			if err := r.neoStore.UpsertNode(ctx, existing); err != nil {
				r.logger.Warn("failed to update neo4j node", zap.Error(err))
			}
		}

		r.publishEvent(ctx, model.ChangeActionUpdated, existing, diff)
	} else {
		// Just update last_seen
		if err := r.pgStore.Update(ctx, existing); err != nil {
			return fmt.Errorf("updating last_seen: %w", err)
		}
	}

	return nil
}

func (r *Reconciler) reconcileRelationship(ctx context.Context, rel *model.Relationship) error {
	if r.neoStore == nil {
		return nil
	}
	return r.neoStore.UpsertRelationship(ctx, rel)
}

func (r *Reconciler) publishEvent(ctx context.Context, action model.ChangeAction, asset *model.Asset, diff map[string]any) {
	if r.publisher == nil {
		return
	}

	diffJSON, _ := json.Marshal(diff)
	event := model.NewChangeEvent(asset.ID, action, asset.Source, diffJSON)

	// Store the change event
	if err := r.pgStore.CreateChangeEvent(ctx, event); err != nil {
		r.logger.Warn("failed to store change event", zap.Error(err))
	}

	// Publish to event bus
	eventData, _ := json.Marshal(event)
	subject := "assets." + string(action)
	if err := r.publisher.Publish(ctx, subject, eventData); err != nil {
		r.logger.Warn("failed to publish event", zap.String("subject", subject), zap.Error(err))
	}
}

// detectChanges compares two assets and returns a map of changed fields.
func detectChanges(existing, incoming *model.Asset) map[string]any {
	diff := make(map[string]any)

	if existing.Name != incoming.Name {
		diff["name"] = map[string]string{"old": existing.Name, "new": incoming.Name}
	}

	existingFQDN := ""
	incomingFQDN := ""
	if existing.FQDN != nil {
		existingFQDN = *existing.FQDN
	}
	if incoming.FQDN != nil {
		incomingFQDN = *incoming.FQDN
	}
	if existingFQDN != incomingFQDN {
		diff["fqdn"] = map[string]string{"old": existingFQDN, "new": incomingFQDN}
	}

	if string(existing.Attributes) != string(incoming.Attributes) {
		diff["attributes"] = map[string]string{
			"old": string(existing.Attributes),
			"new": string(incoming.Attributes),
		}
	}

	return diff
}
