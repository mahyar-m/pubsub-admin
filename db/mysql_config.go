package db

type MysqlConfig struct {
}

func (mc *MysqlConfig) GetConnString() string {
	// TODO: read it from a config file or env variable
	return "root:root@tcp(localhost)/golang-docker?parseTime=true"
}
