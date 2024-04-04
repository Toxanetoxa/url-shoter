package editAlias

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shoter/internal/lib/api/response"
	"url-shoter/internal/lib/logger/sl"
)

type Request struct {
	Alias string `json:"alias,omitempty"`
	ID    int64  `json:"id,omitempty"`
}

type Response struct {
	resp.Response
	Info string `json:"info,omitempty"`
}

type editorAlias interface {
	ReplacementAliasByID(id int64, alias string) (int64, error)
}

func New(log *slog.Logger, editor editorAlias) http.HandlerFunc {
	const op = "internal.http.handlers.url.save.New"

	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Не удалось расшифровать тело запроса", sl.Err(err))

			render.JSON(w, r, resp.Error("не удалось расшифровать запрос"))

			return
		}

		log.Info("тело запроса обработано", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		id := req.ID
		alias := req.Alias

		_, err = editor.ReplacementAliasByID(id, alias)
		if err != nil {
			log.Error("не удалось сменить алиас", sl.Err(err))
			responseError(w, r, "не удалось сменить алиас")
			return
		}

		responseOk(w, r, "success")
	}
}

func responseOk(w http.ResponseWriter, r *http.Request, info string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Info:     info,
	})
}

func responseError(w http.ResponseWriter, r *http.Request, error string) {
	render.JSON(w, r, Response{
		Response: resp.Error(error),
	})
}
