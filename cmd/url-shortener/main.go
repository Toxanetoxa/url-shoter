package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"url-shoter/internal/config"
	"url-shoter/internal/http-server/handlers/url/delete"
	"url-shoter/internal/http-server/handlers/url/redirect"
	"url-shoter/internal/http-server/handlers/url/save"
	"url-shoter/internal/http-server/handlers/url/showAll"
	mwLogger "url-shoter/internal/http-server/middleware/logger"
	"url-shoter/internal/logger"
	"url-shoter/internal/storage/pgsql"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func main() {
	//init config: clean env
	cfg := config.MustLoad()

	//init logger: slog
	var log *slog.Logger
	log = logger.SetupLogger(cfg.Env)
	log.Info(
		"starting url-shooter",
		slog.String("env", cfg.Env),
	)

	//init storage: pgsql
	var pgcfg DBConfig = DBConfig{
		Host:     cfg.PGSQL.DBHost,
		Port:     cfg.PGSQL.DBPort,
		User:     cfg.PGSQL.DBUser,
		Password: cfg.PGSQL.DBPass,
		DBName:   cfg.PGSQL.DBName,
		SSLMode:  cfg.PGSQL.DBSSLMode,
	}

	storage, err := pgsql.ConnectDB(pgsql.DBConfig(pgcfg))
	if err != nil {
		log.Error("Неудалось подключиться к бд:", pgcfg)
		os.Exit(1)
	}

	//err = storage.DeleteById(1)
	//if err != nil {
	//	log.Error("Неудалось подключиться к бд:", pgcfg)
	//	os.Exit(1)
	//}
	//подключение к бд это будет использоваться внутри хендлеров

	//init router: chi, "chi-render"
	router := chi.NewRouter()

	//middleware
	router.Use(middleware.RequestID) // добавляет реквест айди к каждому запросу для трейсинга
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger) // логирует все запросы из минусов свой логгер
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(mwLogger.New(*log)) // кастомный логгер

	//routing breakpoints
	//router.Route("/url", func(r chi.Router) {
	//	r.Use(middleware.BasicAuth("url-shooter"), map[string]string{
	//		user: password,
	//	})
	//})
	//get
	router.Get("/all", showAll.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))
	//post
	router.Post("/url", save.New(log, storage, cfg.AliasLength))
	//delete
	router.Delete("/url/{id}", delete.Delete(log, storage))

	log.Info("сервер запущен", slog.String("address", cfg.Address))

	//run server
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout, // время на прочитать запрос
		WriteTimeout: cfg.HTTPServer.Timeout, // время на ответить на запрос
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Error("Ошибка при загрузке сервера")
	}

	log.Error("сервер остановлен")
}
