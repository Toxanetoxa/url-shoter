package pgsql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log"
	"url-shoter/internal/storage"
)

type Storage struct {
	db *sql.DB
}

type URLData struct {
	Id    int    `json:"id"`
	Alias string `json:"alias"`
	Url   string `json:"url"`
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectDB Подключение к бд
func ConnectDB(configDB DBConfig) (*Storage, error) {
	const op = "storage.pgsql.ConnectDB"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", configDB.Host, configDB.Port, configDB.User, configDB.Password, configDB.DBName, configDB.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("%w :Hе удалось пождключиться к бд с данными: %w", psqlInfo, op)
	}

	log.Println("Подключение к бд прошло успешно")

	_, err = CheckTableExist(db)
	if err != nil {
		log.Fatalf("%w :Hе удалось подключиться к таблице urls", op)
	}

	return &Storage{db: db}, nil
}

// CheckTableExist Проверка наличия таблицы
func CheckTableExist(db *sql.DB) (bool, error) {
	const op = "storage.pgsql.CheckTableExist"
	var tableExists bool

	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'urls')").Scan(&tableExists)
	if err != nil {
		log.Fatalf("%w :failed to check if table exists: %v", op, err)
	}

	if !tableExists {
		_, err = db.Exec(`CREATE TABLE urls (id SERIAL PRIMARY KEY, alias TEXT NOT NULL UNIQUE, url TEXT NOT NULL)`)
		if err != nil {
			return tableExists, fmt.Errorf("%w :Ошибка, таблицы 'urls' нет и не удаётся её создать: %v", op, err)
		}
		log.Println("Таблица 'urls' успешно создана")
		return !tableExists, nil
	} else {
		return tableExists, nil
	}
}

// SaveUrl Сохранение нового url с алиасом
func (s *Storage) SaveUrl(urlToSave string, alias string) (int64, error) {
	const op = "storage.pgsql.SaveUrl"

	stmt, err := s.db.Prepare("INSERT INTO urls(url, alias) VALUES ($1, $2) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s : Неудалось записать значение %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			_ = fmt.Errorf("%s: не удалось закрыть подключение к базе данных", op, err)
		}
	}(stmt)

	var id int64
	err = stmt.QueryRow(urlToSave, alias).Scan(&id)
	if err != nil {
		pgErr, isPGErr := err.(*pq.Error)
		if isPGErr && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// GetURL GetUrl Получение url
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.pgsql.GetUrl"
	stmt, err := s.db.Prepare("SELECT url FROM urls WHERE alias = $1")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: execute statemeny %w", op, err)
	}

	return resURL, nil
}

// TODO написать удаление url
//func (s *Storage) DeleteUrl(alias string) (string, error) {
//}

// CheckAllUrls вывод всех записей из таблицы для дебага
func (s *Storage) CheckAllUrls() ([]URLData, error) {
	const op = "storage.pgsql.CheckAllUrls"
	rows, err := s.db.Query("SELECT * FROM urls")
	if err != nil {
		_ = fmt.Errorf("%s: не удалось получить все записи из базы данных %w", op, err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		var err = rows.Close()
		if err != nil {
			_ = fmt.Errorf("%s: не удалось закрыть подключение к базе данных", op, err)
		}
	}(rows)

	var urlsDataList []URLData

	for rows.Next() {
		var urlData URLData
		err := rows.Scan(&urlData.Id, &urlData.Alias, &urlData.Url)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		urlsDataList = append(urlsDataList, urlData)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urlsDataList, nil
}
