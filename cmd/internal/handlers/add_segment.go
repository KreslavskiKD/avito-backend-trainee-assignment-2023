package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"service_api/cmd/internal/db"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func AddSegment(log *slog.Logger, segmentCreator SegmentCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const errorPath = "handlers.AddSegment"

		log = log.With(
			slog.String("errorPath", errorPath),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		if req.Name == "" {
			log.Error("segment name cannot be empty", slog.String("name", req.Name))
			render.JSON(w, r, MakeResponseError("segment name cannot be empty"))
			return
		}

		err = segmentCreator.CreateSegment(req.Name)
		if errors.Is(err, db.ErrSegmentAlreadyExists) {
			log.Info("segment already exists", slog.String("name", req.Name))
			render.JSON(w, r, MakeResponseError("segment already exists"))
			return
		}

		if err != nil {
			log.Error("failed to create segment", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to create segment"))
			return
		}

		log.Info("segment created successfully")
		render.JSON(w, r, OkResponse())
	}
}
