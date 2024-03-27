package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shoter/internal/lib/api/response"
	"url-shoter/internal/lib/logger/sl"
	"url-shoter/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.42.1 --name=URLGetter

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http.handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("%s, alias ек обнаружен", op)

			render.JSON(w, r, resp.Error("не найдено"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url не обнаружен", "alias", alias)

			render.JSON(w, r, resp.Error("не обнаружено"))

			return
		}
		if err != nil {
			log.Error("не удалось создать URL", sl.Err(err))

			render.JSON(w, r, resp.Error("Другая ошибка"))

			return
		}

		log.Info("редирект по урлу", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
