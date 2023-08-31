package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func DeleteSegment(log *slog.Logger, segmentDeleter SegmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const errorPath = "handlers.DeleteSegment"
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
		err = segmentDeleter.DeleteSegment(req.Name)
		if err != nil {
			log.Error("failed to delete segment", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to delete segment"))
			return
		}

		log.Info("segment deleted successfully")
		render.JSON(w, r, OkResponse())
	}
}
