package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/qrunner/arch/internal/model"
	"github.com/qrunner/arch/internal/store"
	"go.uber.org/zap"
)

// --- Response helpers ---

type apiResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Total int    `json:"total,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiResponse{Error: msg})
}

func parseUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid UUID: "+idStr)
		return uuid.Nil, false
	}
	return id, true
}

func paginationParams(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// --- Health ---

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Assets ---

func (s *Server) handleListAssets(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginationParams(r)
	filter := store.AssetFilter{
		Source:    r.URL.Query().Get("source"),
		AssetType: r.URL.Query().Get("asset_type"),
		Status:    r.URL.Query().Get("status"),
		Search:    r.URL.Query().Get("search"),
		Limit:     limit,
		Offset:    offset,
	}

	assets, total, err := s.pgStore.List(r.Context(), filter)
	if err != nil {
		s.logger.Error("listing assets", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list assets")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: assets, Total: total})
}

func (s *Server) handleGetAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	asset, err := s.pgStore.GetByID(r.Context(), id)
	if err != nil {
		s.logger.Error("getting asset", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get asset")
		return
	}
	if asset == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: asset})
}

func (s *Server) handleCreateAsset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ExternalID  string          `json:"external_id"`
		Source      string          `json:"source"`
		AssetType   string          `json:"asset_type"`
		Name        string          `json:"name"`
		FQDN        *string         `json:"fqdn"`
		IPAddresses []string        `json:"ip_addresses"`
		Attributes  json.RawMessage `json:"attributes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.ExternalID == "" || input.Source == "" || input.AssetType == "" || input.Name == "" {
		writeError(w, http.StatusBadRequest, "external_id, source, asset_type, and name are required")
		return
	}

	asset := model.NewAsset(input.ExternalID, input.Source, input.AssetType, input.Name)
	asset.FQDN = input.FQDN
	if input.IPAddresses != nil {
		asset.IPAddresses = input.IPAddresses
	}
	if input.Attributes != nil {
		asset.Attributes = input.Attributes
	}

	if err := s.pgStore.Create(r.Context(), asset); err != nil {
		s.logger.Error("creating asset", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create asset")
		return
	}

	// Also create a node in Neo4j
	if s.neoStore != nil {
		if err := s.neoStore.UpsertNode(r.Context(), asset); err != nil {
			s.logger.Warn("failed to upsert neo4j node", zap.Error(err))
		}
	}

	writeJSON(w, http.StatusCreated, apiResponse{Data: asset})
}

func (s *Server) handleUpdateAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	existing, err := s.pgStore.GetByID(r.Context(), id)
	if err != nil {
		s.logger.Error("getting asset for update", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get asset")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}

	var input struct {
		Name        *string         `json:"name"`
		FQDN        *string         `json:"fqdn"`
		IPAddresses []string        `json:"ip_addresses"`
		Attributes  json.RawMessage `json:"attributes"`
		Status      *string         `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.FQDN != nil {
		existing.FQDN = input.FQDN
	}
	if input.IPAddresses != nil {
		existing.IPAddresses = input.IPAddresses
	}
	if input.Attributes != nil {
		existing.Attributes = input.Attributes
	}
	if input.Status != nil {
		existing.Status = model.AssetStatus(*input.Status)
	}

	if err := s.pgStore.Update(r.Context(), existing); err != nil {
		s.logger.Error("updating asset", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to update asset")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: existing})
}

func (s *Server) handleDeleteAsset(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	if err := s.pgStore.Delete(r.Context(), id); err != nil {
		s.logger.Error("deleting asset", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to delete asset")
		return
	}

	if s.neoStore != nil {
		if err := s.neoStore.DeleteNode(r.Context(), id); err != nil {
			s.logger.Warn("failed to delete neo4j node", zap.Error(err))
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- History ---

func (s *Server) handleGetAssetHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	limit, offset := paginationParams(r)
	events, total, err := s.pgStore.ListChangeEventsByAssetID(r.Context(), id, limit, offset)
	if err != nil {
		s.logger.Error("listing asset history", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list history")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: events, Total: total})
}

// --- Relationships ---

func (s *Server) handleGetAssetRelationships(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	if s.neoStore == nil {
		writeError(w, http.StatusServiceUnavailable, "graph store not available")
		return
	}

	rels, err := s.neoStore.GetRelationships(r.Context(), id)
	if err != nil {
		s.logger.Error("getting relationships", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get relationships")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: rels})
}

// --- Graph ---

func (s *Server) handleGetDependencyGraph(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	if s.neoStore == nil {
		writeError(w, http.StatusServiceUnavailable, "graph store not available")
		return
	}

	depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
	if depth <= 0 {
		depth = 3
	}

	assets, rels, err := s.neoStore.GetDependencyGraph(r.Context(), id, depth)
	if err != nil {
		s.logger.Error("getting dependency graph", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get dependency graph")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: map[string]any{
		"assets":        assets,
		"relationships": rels,
	}})
}

func (s *Server) handleGetImpactGraph(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r)
	if !ok {
		return
	}

	if s.neoStore == nil {
		writeError(w, http.StatusServiceUnavailable, "graph store not available")
		return
	}

	depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
	if depth <= 0 {
		depth = 3
	}

	assets, rels, err := s.neoStore.GetImpactGraph(r.Context(), id, depth)
	if err != nil {
		s.logger.Error("getting impact graph", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to get impact graph")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: map[string]any{
		"assets":        assets,
		"relationships": rels,
	}})
}

// --- Collectors ---

func (s *Server) handleListCollectors(w http.ResponseWriter, r *http.Request) {
	// Placeholder - will be populated when collector registry is wired in
	writeJSON(w, http.StatusOK, apiResponse{Data: []any{}})
}

func (s *Server) handleTriggerCollector(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	// Placeholder - will trigger actual collection when wired in
	writeJSON(w, http.StatusAccepted, apiResponse{Data: map[string]string{
		"message":   "collection triggered",
		"collector": name,
	}})
}

// --- Changes ---

func (s *Server) handleListChanges(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginationParams(r)
	events, total, err := s.pgStore.ListRecentChangeEvents(r.Context(), limit, offset)
	if err != nil {
		s.logger.Error("listing changes", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list changes")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Data: events, Total: total})
}

// --- Dashboard ---

func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Basic stats query
	ctx := r.Context()

	type stats struct {
		TotalAssets  int            `json:"total_assets"`
		BySource     map[string]int `json:"by_source"`
		ByType       map[string]int `json:"by_type"`
		ByStatus     map[string]int `json:"by_status"`
		RecentChanges int           `json:"recent_changes"`
	}

	st := stats{
		BySource: make(map[string]int),
		ByType:   make(map[string]int),
		ByStatus: make(map[string]int),
	}

	// Total count
	pool := s.pgStore.Pool()
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM assets").Scan(&st.TotalAssets)

	// By source
	rows, err := pool.Query(ctx, "SELECT source, COUNT(*) FROM assets GROUP BY source")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var src string
			var cnt int
			rows.Scan(&src, &cnt)
			st.BySource[src] = cnt
		}
	}

	// By type
	rows2, err := pool.Query(ctx, "SELECT asset_type, COUNT(*) FROM assets GROUP BY asset_type")
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var at string
			var cnt int
			rows2.Scan(&at, &cnt)
			st.ByType[at] = cnt
		}
	}

	// By status
	rows3, err := pool.Query(ctx, "SELECT status, COUNT(*) FROM assets GROUP BY status")
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var status string
			var cnt int
			rows3.Scan(&status, &cnt)
			st.ByStatus[status] = cnt
		}
	}

	// Recent changes count (last 24h)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM change_events WHERE timestamp > NOW() - INTERVAL '24 hours'").Scan(&st.RecentChanges)

	writeJSON(w, http.StatusOK, apiResponse{Data: st})
}

// --- SSE Events ---

func (s *Server) handleSSEEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// Send initial connection event
	w.Write([]byte("event: connected\ndata: {\"status\":\"connected\"}\n\n"))
	flusher.Flush()

	// Keep connection open until client disconnects.
	// In production, this will subscribe to NATS and forward events.
	<-r.Context().Done()
}
