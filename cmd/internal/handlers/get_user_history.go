package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"service_api/cmd/internal/config"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type RequestUH struct {
	UserId int64      `json:"userId"`
	Year   int        `json:"year"`
	Month  time.Month `json:"month"`
}

type ResponseUH struct {
	Response
	Link string `json:"link"`
}

func GetUserHistory(log *slog.Logger, historyGetter HistoryPreparator, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetUserHistory"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req RequestUH
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		err = historyGetter.PrepareUserHistoryFile(req.UserId, req.Year, req.Month)

		if err != nil {
			log.Error("failed to get user history", MakeLoggerError(err))
			render.JSON(w, r, MakeResponseError("failed to get user history"))
			return
		}

		log.Info("user history retrieved successfully")
		link := cfg.Address + "/report/" + fmt.Sprintf("%d", req.UserId)
		render.JSON(w, r, ResponseUH{
			OkResponse(),
			link,
		})
	}

}
