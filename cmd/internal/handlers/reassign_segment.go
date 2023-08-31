package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"service_api/cmd/internal/db"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type RequestREAS struct {
	SegmentsToAdd    []string `json:"segmentsToAdd"`
	SegmentsToRemove []string `json:"segmentsToRemove"`
	UserId           int64    `json:"userId"`
}

func ReassignSegments(log *slog.Logger, segmentAssigner SegmentAssigner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.ReassignSegments"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req RequestREAS
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if len(req.SegmentsToRemove) == 0 && len(req.SegmentsToAdd) == 0 {
			log.Error("no segments were given")
			render.JSON(w, r, MakeResponseError("no segments were given"))
			return
		}

		err = segmentAssigner.ReassignSegments(req.SegmentsToAdd, req.SegmentsToRemove, req.UserId)
		if errors.Is(err, db.ErrUserAlreadyHasSegment) {
			log.Info("user already has some of the segments", slog.String("segmentsToAdd", strings.Join(req.SegmentsToAdd, ",")))
			render.JSON(w, r, MakeResponseError("user already has some of the segments"))
			return
		}
		if errors.Is(err, db.ErrUserNotFound) {
			log.Info("user not found", slog.Int64("userId", req.UserId))
			render.JSON(w, r, MakeResponseError("user not found"))
			return
		}
		if err != nil {
			log.Error("failed to reassign segments", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to reassign segments"))
			return
		}
		log.Info("segments reassigned successfully")
		render.JSON(w, r, OkResponse())
	}
}
