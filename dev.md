cd f:\MY_TEST\nps
go run cmd/nps/nps.go



# 指定配置文件目录（用其他目录的 conf）
go run cmd/nps/nps.go -conf_path=f:\MY_TEST\nps\conf

# 指定日志输出文件
go run cmd/nps/nps.go -log_path=f:\MY_TEST\nps\nps.log




go run cmd/npc/npc.go -server=127.0.0.1:8024 -vkey=123 -type=tcp