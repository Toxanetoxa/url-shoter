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
	Id    int64  `json:"id"`
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

/*
ConnectDB Подключение к бд
*/
func ConnectDB(configDB DBConfig) (*Storage, error) {
	const op = "storage.pgsql.ConnectDB"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", configDB.Host, configDB.Port, configDB.User, configDB.Password, configDB.DBName, configDB.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("%w :Hе удалось подключиться к бд с данными: %w", psqlInfo, op)
	}

	log.Println("Подключение к бд прошло успешно")

	_, err = CheckTableExist(db)
	if err != nil {
		log.Fatalf("%w :Hе удалось подключиться к таблице urls", op)
	}

	return &Storage{db: db}, nil
}

/*
CheckTableExist Проверка наличия таблицы
*/
func CheckTableExist(db *sql.DB) (bool, error) {
	const op = "storage.pgsql.CheckTableExist"
	var tableExists bool

	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'urls')").Scan(&tableExists)
	if err != nil {
		_ = fmt.Errorf("%s :ошибка, таблица была не обнаружена: %v", op, err)
	}

	if !tableExists {
		_, err = db.Exec(`CREATE TABLE urls (id SERIAL PRIMARY KEY, alias TEXT NOT NULL UNIQUE, url TEXT NOT NULL)`)
		if err != nil {
			return tableExists, fmt.Errorf("%w :Ошибка, таблицы 'urls' не удаётся её создать: %v", op, err)
		}
		log.Println("Таблица 'urls' успешно создана")
		return !tableExists, nil
	} else {
		return tableExists, nil
	}
}

/*
SaveUrl Сохранение нового url с алиасом и id не обязательный параметр
*/
func (s *Storage) SaveUrl(urlToSave string, alias string, id *int64) (int64, error) {
	const op = "storage.pgsql.SaveUrl"

	var stmt *sql.Stmt
	var err error

	if id != nil {
		isUrl, err := s.ExistUrlById(*id)
		if isUrl || err != nil {
			return 0, fmt.Errorf("%s: %w\n", op, err)
		}
		stmt, err = s.db.Prepare("INSERT INTO urls(url, alias, id) VALUES ($1, $2, $3) RETURNING id")
	} else {
		stmt, err = s.db.Prepare("INSERT INTO urls(url, alias) VALUES ($1, $2) RETURNING id")
	}

	if err != nil {
		return 0, fmt.Errorf("%s : Неудалось записать значение %w\n", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			LogErrorCloseDb(op, err)
		}
	}(stmt)

	var newID int64
	if id != nil {
		err = stmt.QueryRow(urlToSave, alias, id).Scan(&newID)
	} else {
		err = stmt.QueryRow(urlToSave, alias).Scan(&newID)
	}
	if err != nil {
		pgErr, isPGErr := err.(*pq.Error)
		if isPGErr && pgErr.Code == "23505" {
			return 0, fmt.Errorf("%s", op, err)
		}
		return 0, fmt.Errorf("%s", op, err)
	}

	return newID, nil
}

/*
GetURL Получение url
*/
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

/*
GetUrlById Получение url  по id
*/
func (s *Storage) GetUrlById(id int64) (string, error) {
	const op = "storage.pgsql.GetUrlById"

	stmt, err := s.db.Prepare("SELECT url FROM urls WHERE id = $1")
	if err != nil {
		return "", fmt.Errorf("%s: не удалось подклбчиться к бд: %w", op, err)
	}

	var resURL string
	err = stmt.QueryRow(id).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: не удалось выполнить скрипт %w", op, err)
	}

	return resURL, nil
}

/*
DeleteById удаление урла из таблицы по id
*/
func (s *Storage) DeleteById(id int64) error {
	const op = "storage.pgsql.DeleteById"

	stmt, err := s.db.Prepare("DELETE FROM urls WHERE id=$1")
	if err != nil {
		return fmt.Errorf("%s: не удалоcь удалить url по id: %w", op, id, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			LogErrorCloseDb(op, err)
		}
	}(stmt)

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: не удалось выполнить запрос на удаление URL по ID: %w", op, id, err)
	}

	log.Printf("Удаление url по id :%s прошло успешно", id)
	return nil
}

/*
ExistUrlById проверяет наличие записи URL по ID.
Если запись существует, возвращает true и nil.
Если запись не существует, возвращает false и nil.
*/
func (s *Storage) ExistUrlById(id int64) (bool, error) {
	const op = "storage.pgsql.ExistUrlById"
	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM urls WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s: не удалось подготовить запрос на поиск URL по ID: %v", op, id, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			LogErrorCloseDb(op, err)
		}
	}(stmt)

	var count int64
	err = stmt.QueryRow(id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("%s: не удалось выполнить запрос на поиск URL по ID: %v", op, id, err)
	}

	// Если количество записей с указанным ID больше нуля, то URL существует
	return count > 0, nil
}

/*
ExistUrlByAlias проверка наличия урла по алиасу
*/
func (s *Storage) ExistUrlByAlias(alias string) (bool, error) {
	const op = "storage.pgsql.ExistUrlByAlias"
	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM urls WHERE alias = $1")
	if err != nil {
		return false, fmt.Errorf("%s: не удалось подготовить запрос на поиск URL по Alias: %v", op, alias, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			LogErrorCloseDb(op, err)
		}
	}(stmt)

	var count int64
	err = stmt.QueryRow(alias).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("%s: не удалось выполнить запрос на поиск URL по Alias: %v", op, err)
	}

	return count > 0, nil
}

/*
CheckAllUrls вывод всех записей из таблицы для дебага
*/
func (s *Storage) CheckAllUrls() ([]URLData, error) {
	const op = "storage.pgsql.CheckAllUrls"
	rows, err := s.db.Query("SELECT * FROM urls")
	if err != nil {
		return nil, fmt.Errorf("%s: не удалось получить все записи из базы данных: %v", op, err)
	}
	defer func(rows *sql.Rows) {
		var err = rows.Close()
		if err != nil {
			LogErrorCloseDb(op, err)
		}
	}(rows)

	var urlsDataList []URLData

	for rows.Next() {
		var urlData URLData
		err := rows.Scan(&urlData.Id, &urlData.Alias, &urlData.Url)
		if err != nil {
			return nil, fmt.Errorf("%s, не удалось выполнить скрипт на вывод всех записей из таблицы: %v", op, err)
		}
		urlsDataList = append(urlsDataList, urlData)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s, данные в таблице urls не были обнаружены: %v", op, err)
	}
	return urlsDataList, nil
}

/*
ReplacementAliasByID удаление записи по id и добавление новой записи с новым алиасом
*/
func (s *Storage) ReplacementAliasByID(id int64, alias string) (int64, error) {
	const op = "storage.pgsql.ReplacementAliasByID"

	isUrl, err := s.ExistUrlById(id)
	if !isUrl || err != nil {
		return 0, fmt.Errorf("%s, урл по указанному ID: %d был не обнаружен: %v", op, id, err)
	}

	url, err := s.GetURL(alias)
	if url != "" || err == nil {
		return 0, fmt.Errorf("%s, указанный alias:%v уже занят: %v", op, alias, err)
	}

	saveUrl, err := s.GetUrlById(id)
	if err != nil {
		return 0, fmt.Errorf("%s, урл по указанному айди был не обнаружен: %v", op, err)
	}

	err = s.DeleteById(id)
	if err != nil {
		return 0, fmt.Errorf("%s, не удалось удалить урл по старому айдишнику: %v", op, err)
	}

	_, err = s.SaveUrl(saveUrl, alias, &id)
	if err != nil {
		return 0, fmt.Errorf("%s, не удалось сохранить урл с новым алиасом: %v", op, err)
	}

	return id, nil
}

/*
LogErrorCloseDb функция хелпер вывода лога ошибки неудачного закрытия соединения с бд
*/
func LogErrorCloseDb(op string, err error) {
	_ = fmt.Errorf("%s, не удалось отключиться от базы данных: %v", op, err)
}
