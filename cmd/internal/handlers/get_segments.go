package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type RequestSG struct {
	UserId int64 `json:"userId"`
}

type ResponseSG struct {
	Response
	Segments []string `json:"segments"`
}

func GetSegments(log *slog.Logger, segmentGetter SegmentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const errorPath = "handlers.GetSegments"
		log = log.With(
			slog.String("errorPath", errorPath),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req RequestSG
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		segments, err := segmentGetter.GetSegments(req.UserId)
		if err != nil {
			log.Error("failed to get segments", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to get segments"))
			return
		}

		log.Info("segments retrieved successfully")
		render.JSON(w, r, ResponseSG{
			OkResponse(),
			segments,
		})
	}
}
