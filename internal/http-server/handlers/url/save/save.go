package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shoter/internal/lib/api/response"
	"url-shoter/internal/lib/logger/sl"
	"url-shoter/internal/lib/random"
	"url-shoter/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveUrl(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver, aliasLength int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http.handlers.url.save.New"

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

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
			// реализовать проверку на уже существующий алиас
		}

		id, err := urlSaver.SaveUrl(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("URL уже существует", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url уже существует"))

			return
		}
		if err != nil {
			log.Error("не удалось создать URL", sl.Err(err))

			render.JSON(w, r, resp.Error("не удалось создать URL"))

			return
		}

		log.Info("добавлен url", slog.Int64("id", id))

		responseOk(w, r, alias)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
