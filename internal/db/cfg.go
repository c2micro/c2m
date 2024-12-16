package db

// стркуктура для кофигурации БД
type ConfigV1 struct {
	Sqlite     *ConfigSqliteV1     `json:"sqlite" validate:"omitempty"`
	Postgresql *ConfigPostgresqlV1 `json:"postgresql" validate:"omitempty"`
}
