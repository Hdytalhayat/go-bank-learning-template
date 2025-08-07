package config

const dsn = "root@127.0.0.1:3306/bank_app_db?parseTime=true"

// dsn = "user:password@tcp(host:port)/database_name?param1=value1¶m2=value2"
// parseTime=true diperlukan agar tipe data MySQL DATETIME/TIMESTAMP bisa di-scan ke tipe time.Time Go.
