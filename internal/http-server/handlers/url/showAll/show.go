package showAll

import (
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shoter/internal/lib/api/response"
	"url-shoter/internal/storage/pgsql"
)

type URLsViewer interface {
	CheckAllUrls() (pgsql.URLData, error)
}

type Response struct {
	resp.Response
	List  []pgsql.URLData `json:"list,omitempty"`
	Count int             `json:"count,omitempty"`
}

func New(log *slog.Logger, viewer *pgsql.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http.handlers.url.showAll.New"

		log = log.With(
			slog.String("op", op),
		)

		urls, err := showAllUrls(w, r, viewer)
		if err != nil {
			responseErr(w, r)
		}

		responseOk(w, r, urls)
	}
}

func showAllUrls(w http.ResponseWriter, r *http.Request, viewer *pgsql.Storage) ([]pgsql.URLData, error) {
	const op = "internal.http.handlers.url.showAll.showAllUrls"

	urls, err := viewer.CheckAllUrls()
	if err != nil {
		slog.String("%s, Данные по урлам не обнаружены", op)
		return nil, err
	}

	return urls, nil

}

func responseOk(w http.ResponseWriter, r *http.Request, data []pgsql.URLData) {

	var urlDataList []pgsql.URLData
	for _, i := range data {
		urlDataList = append(urlDataList, pgsql.URLData{
			Id:    i.Id,
			Alias: i.Alias,
			Url:   i.Url,
		})
	}

	render.JSON(w, r, Response{
		Response: resp.OK(),
		List:     urlDataList,
		Count:    len(urlDataList),
	})
}

func responseErr(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.Error("Данные по урлам не обнаружены"),
		List:     nil,
		Count:    0,
	})
}
