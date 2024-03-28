package delete

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	resp "url-shoter/internal/lib/api/response"
)

type UrlDeleter interface {
	DeleteById(id int64) error
	ExistUrlById(id int64) (bool, error)
}

type Request struct {
	Id int64 `json:"id"`
}

type Response struct {
	resp.Response
	Error error `json:"error,omitempty"`
}

func Delete(log *slog.Logger, deleter UrlDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http.handlers.url.delete.Delete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			log.Error(op, "не верно передан id, он должен быть типа int64", id, err)
			responseError(w, r, err)
		}

		exist, err := deleter.ExistUrlById(id)
		if err != nil || !exist {
			log.Error(op, "не удалось найти урл с соотвествующим id", id, err)
			responseError(w, r, err)
			return
		}

		err = deleter.DeleteById(id)
		if err != nil {
			log.Error(op, "Не удалось удалить url по id", id, err)
			responseError(w, r, err)
			return
		}

		responseOk(w, r)

	}
}

func responseOk(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Error:    nil,
	})
}

func responseError(w http.ResponseWriter, r *http.Request, err error) {
	render.JSON(w, r, Response{
		Response: resp.Error("id not exist"),
		Error:    err,
	})
}
