package db

// TODO структура для postgresql драйвера
type ConfigPostgresqlV1 struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	User     string `json:"user" validate:"required"`
	Db       string `json:"db" validate:"required"`
	Password string `json:"password" validate:"required"`
}
