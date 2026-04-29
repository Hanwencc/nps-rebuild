package main

import (
	"bufio"
	"ehang.io/nps/bridge"
	"ehang.io/nps/lib/daemon"
	"ehang.io/nps/server"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/file/sqlitedb"
	"ehang.io/nps/lib/install"
	"ehang.io/nps/lib/version"
	"ehang.io/nps/server/connection"
	"ehang.io/nps/server/tool"
	"ehang.io/nps/web/routers"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"github.com/kardianos/service"
)

var (
	level      string
	ver        = flag.Bool("version", false, "show current version")
	confPath   = flag.String("conf_path", "", "set current confPath")
	serverCmd  = flag.Bool("server", false, "NPS管理脚本")
	npsLogPath = flag.String("log_path", "", "nps log path")
)

func main() {

	debug.SetMaxThreads(1000000)

	flag.Parse()
	// init log
	if *ver {
		common.PrintVersion()
		return
	}
	if *serverCmd {
		printSlogan()
		inputCmd()
		return
	}

	var logPath string
	// *confPath why get null value ?
	for _, v := range os.Args[1:] {
		switch v {
		case "install", "start", "stop", "uninstall", "restart":
			continue
		}
		if strings.Contains(v, "-conf_path=") {
			common.ConfPath = strings.Replace(v, "-conf_path=", "", -1)
		}

		if strings.Contains(v, "-log_path=") {
			logPath = strings.Replace(v, "-log_path=", "", -1)
		}
	}

	if err := beego.LoadAppConfig("ini", filepath.Join(common.GetRunPath(), "conf", "nps.conf")); err != nil {
		log.Fatalln("load config file error", err.Error())
	}

	common.InitPProfFromFile()
	if level = beego.AppConfig.String("log_level"); level == "" {
		level = "7"
	}
	logs.Reset()
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)

	if logPath == "" {
		logPath := beego.AppConfig.String("log_path")
		if logPath == "" {
			logPath = common.GetLogPath()
		}
		if common.IsWindows() {
			logPath = strings.Replace(logPath, "\\", "\\\\", -1)
		}
	}

	// init service
	options := make(service.KeyValue)
	svcConfig := &service.Config{
		Name:        "Nps",
		DisplayName: "nps内网穿透代理服务器",
		Description: "一款轻量级、功能强大的内网穿透代理服务器。支持tcp、udp流量转发，支持内网http代理、内网socks5代理，同时支持snappy压缩、站点保护、加密传输、多路复用、header修改等。支持web图形化管理，集成多用户模式。",
		Option:      options,
	}

	bridge.ServerTlsEnable = beego.AppConfig.DefaultBool("tls_enable", false)

	for _, v := range os.Args[1:] {
		switch v {
		case "install", "start", "stop", "uninstall", "restart":
			continue
		}
		svcConfig.Arguments = append(svcConfig.Arguments, v)
	}

	svcConfig.Arguments = append(svcConfig.Arguments, "service")
	if len(os.Args) > 1 && os.Args[1] == "service" {
		_ = logs.SetLogger(logs.AdapterFile, `{"level":`+level+`,"filename":"`+logPath+`","daily":false,"maxlines":100000,"color":true}`)
	} else {
		_ = logs.SetLogger(logs.AdapterConsole, `{"level":`+level+`,"color":true}`)
	}
	if !common.IsWindows() {
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"}
		svcConfig.Option["SystemdScript"] = install.SystemdScript
		svcConfig.Option["SysvScript"] = install.SysvScript
	}
	prg := &nps{}
	prg.exit = make(chan struct{})
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logs.Error(err, "service function disabled")
		run()
		// run without service
		wg := sync.WaitGroup{}
		wg.Add(1)
		wg.Wait()
		return
	}

	if len(os.Args) > 1 && os.Args[1] != "service" {
		switch os.Args[1] {
		case "reload":
			daemon.InitDaemon("nps", common.GetRunPath(), common.GetTmpPath())
			return
		case "install":
			// uninstall before
			_ = service.Control(s, "stop")
			_ = service.Control(s, "uninstall")

			binPath := install.InstallNps()
			svcConfig.Executable = binPath
			s, err := service.New(prg, svcConfig)
			if err != nil {
				logs.Error(err)
				return
			}
			err = service.Control(s, os.Args[1])
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			}
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				confPath := "/etc/init.d/" + svcConfig.Name
				os.Symlink(confPath, "/etc/rc.d/S90"+svcConfig.Name)
				os.Symlink(confPath, "/etc/rc.d/K02"+svcConfig.Name)
			}
			return
		case "start", "restart", "stop":
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				cmd := exec.Command("/etc/init.d/"+svcConfig.Name, os.Args[1])
				err := cmd.Run()
				if err != nil {
					logs.Error(err)
				}
				return
			}
			err := service.Control(s, os.Args[1])
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			}
			return
		case "uninstall":
			err := service.Control(s, os.Args[1])
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			}
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				os.Remove("/etc/rc.d/S90" + svcConfig.Name)
				os.Remove("/etc/rc.d/K02" + svcConfig.Name)
			}
			return
		case "update":
			install.UpdateNps()
			return
			//default:
			//	logs.Error("command is not support")
			//	return
		}
	}

	_ = s.Run()
}

func printSlogan() {
	green := color.New(color.FgGreen).SprintFunc()
	// 第一次输入，如果输入 1,2,3，4 则需要输入秘钥，否则

	fmt.Printf("%s", green(""))

	fmt.Printf("\033[32;0m欢迎使用 NPS 管理脚本 \n")
	fmt.Printf("\033[0m") // 重置颜色

	fmt.Printf("\n")

	fmt.Printf("\u001B[32m输入[1]\u001B[0m - 安装 NPS\n")
	fmt.Printf("\u001B[32m输入[2]\u001B[0m - 卸载 NPS\n")
	fmt.Printf("\u001B[32m输入[3]\u001B[0m - 更新 NPS\n")
	fmt.Printf("---------------------\n")
	fmt.Printf("\u001B[32m输入[4]\u001B[0m - 查看状态\n")
	fmt.Printf("---------------------\n")
	fmt.Printf("\u001B[32m输入[5]\u001B[0m - 启动 NPS\n")
	fmt.Printf("\u001B[32m输入[6]\u001B[0m - 停止 NPS\n")
	fmt.Printf("\u001B[32m输入[7]\u001B[0m - 重启 NPS\n")
	fmt.Printf("---------------------\n")
	fmt.Printf("\u001B[32m输入[0]\u001B[0m - 退出脚本\n")
	fmt.Printf("---------------------\n")
	fmt.Printf("\n")

}

func inputCmd() {
	var flag string
	fmt.Printf("请输入[0-7]：")

	stdin := bufio.NewReader(os.Stdin)
	_, err := fmt.Fscanln(stdin, &flag)
	if err != nil {
		fmt.Println("输入有误")
	} else {
		if flag == "0" {
			os.Exit(0)
		}

		// init service

		prg := &nps{
			exit: make(chan struct{}),
		}
		options := make(service.KeyValue)
		svcConfig := &service.Config{
			Name:        "Nps",
			DisplayName: "nps内网穿透代理服务器",
			Description: "一款轻量级、功能强大的内网穿透代理服务器。支持tcp、udp流量转发，支持内网http代理、内网socks5代理，同时支持snappy压缩、站点保护、加密传输、多路复用、header修改等。支持web图形化管理，集成多用户模式。",
			Option:      options,
		}
		s, _ := service.New(prg, svcConfig)

		switch flag {
		case "1":
			// uninstall before
			_ = service.Control(s, "stop")
			_ = service.Control(s, "uninstall")
			binPath := install.InstallNpsToCurrentDir()

			beego.LoadAppConfig("ini", filepath.Join(common.GetAppPath(), "conf", "nps.conf"))

			logPath := filepath.Join(common.GetAppPath(), "nps.log")
			if common.IsWindows() {
				logPath = strings.Replace(logPath, "\\", "\\\\", -1)
			}
			svcConfig.Arguments = append(svcConfig.Arguments, "service")
			svcConfig.Arguments = append(svcConfig.Arguments, "-conf_path="+common.GetAppPath())
			svcConfig.Arguments = append(svcConfig.Arguments, "-log_path="+logPath)

			fmt.Println("日志文件路径为：", logPath)

			svcConfig.Executable = binPath
			s, err := service.New(prg, svcConfig)

			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				confPath := "/etc/init.d/" + svcConfig.Name
				os.Symlink(confPath, "/etc/rc.d/S90"+svcConfig.Name)
				os.Symlink(confPath, "/etc/rc.d/K02"+svcConfig.Name)
			}

			err = service.Control(s, "install")
			if err != nil {
				logs.Error("Valid actions: %q\n%s", service.ControlAction, err.Error())
			} else {
				fmt.Println("NPS服务安装成功")
			}

			err = service.Control(s, "start")
			if err != nil {
				fmt.Println("启动NPS服务失败", err)
			} else {
				fmt.Println("NPS服务已启动，管理面板访问地址：127.0.0.1:" + beego.AppConfig.String("web_port"))
			}

			break
		case "2":
			// 卸载系统服务
			err := service.Control(s, "stop")
			if err != nil {
				fmt.Println("NPS服务停止失败", err)
			} else {
				fmt.Println("NPS服务已停止")
			}

			err = service.Control(s, "uninstall")
			if err != nil {
				logs.Error("NPS服务卸载失败")
			}
			if service.Platform() == "unix-systemv" {
				logs.Info("unix-systemv service")
				os.Remove("/etc/rc.d/S90" + svcConfig.Name)
				os.Remove("/etc/rc.d/K02" + svcConfig.Name)
			}

			if err == nil {
				fmt.Println("NPS服务已卸载成功")
			}
			break
		case "3":
			install.UpdateNpsNew()
			return
		case "4":
			// 查看状态
			var statusMsg = ""
			status, err := s.Status()
			if err != nil {
				statusMsg = "\u001B[31m未运行\u001B[0m"
			} else {
				if status == 1 {
					statusMsg = "\u001B[32m运行中\u001B[0m"
				} else {
					statusMsg = "\u001B[31m未运行\u001B[0m"
				}
			}
			fmt.Println("NPS服务状态：" + statusMsg)
			break
		case "5":
			// 启动 NPS
			err := service.Control(s, "start")
			if err != nil {
				fmt.Println("NPS服务启动失败", err)
			} else {
				fmt.Println("NPS服务启动成功")
			}

			break
		case "6":
			// 停止 NPS
			err := service.Control(s, "stop")
			if err != nil {
				fmt.Println("NPS服务停止失败", err)
			} else {
				fmt.Println("NPS服务停止成功")
			}

			break
		case "7":
			// 重启 NPS
			err := service.Control(s, "restart")
			if err != nil {
				fmt.Println("NPS服务重启失败", err)
			} else {
				fmt.Println("NPS服务重启成功")
			}

			break
		}
	}

	inputCmd()
}

func installNps() {

}

type nps struct {
	exit chan struct{}
}

func (p *nps) Start(s service.Service) error {
	_, _ = s.Status()
	go p.run()
	return nil
}
func (p *nps) Stop(s service.Service) error {
	_, _ = s.Status()
	close(p.exit)
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (p *nps) run() error {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logs.Warning("nps: panic serving %v: %v\n%s", err, string(buf))
		}
	}()
	run()
	select {
	case <-p.exit:
		logs.Warning("stop...")
	}
	return nil
}

func run() {
	routers.Init()
	// Phase 0 of JSON->SQLite migration: open the embedded SQLite store.
	// Done here (server entrypoint) rather than in lib/file so npc does
	// not link in modernc.org/sqlite + libc.
	if store, err := sqlitedb.Open(file.GetDb().JsonDb.SqliteFilePath); err != nil {
		logs.Error("open sqlite store at %s failed: %v", file.GetDb().JsonDb.SqliteFilePath, err)
	} else {
		file.GetDb().JsonDb.SQLite = store
		// Phase 1: api_tokens — backfill from api_tokens.json on first
		// boot. Subsequent boots are no-ops (table not empty).
		if n, err := store.BackfillApiTokens(file.GetDb().ListApiTokens()); err != nil {
			logs.Error("backfill api_tokens failed: %v", err)
		} else if n > 0 {
			logs.Info("api_tokens: migrated %d row(s) from api_tokens.json into SQLite", n)
		}
		// Phase 2: clients — either backfill from clients.json on first
		// boot, or REPLACE the in-memory map with rows from SQLite on
		// subsequent boots (SQLite is source of truth).
		if mode, n, err := store.BackfillOrLoadClients(file.GetDb().JsonDb); err != nil {
			logs.Error("clients %s failed: %v", mode, err)
		} else {
			switch mode {
			case "backfill":
				logs.Info("clients: migrated %d row(s) from clients.json into SQLite", n)
			case "load":
				logs.Info("clients: loaded %d row(s) from SQLite", n)
			}
		}
		// Phase 3: tasks — same contract as clients. MUST run AFTER
		// clients so the Client pointer can be resolved by id.
		if mode, n, err := store.BackfillOrLoadTasks(file.GetDb().JsonDb); err != nil {
			logs.Error("tasks %s failed: %v", mode, err)
		} else {
			switch mode {
			case "backfill":
				logs.Info("tasks: migrated %d row(s) from tasks.json into SQLite", n)
			case "load":
				logs.Info("tasks: loaded %d row(s) from SQLite", n)
			}
		}
		// Phase 4: hosts — same contract; also depends on clients.
		if mode, n, err := store.BackfillOrLoadHosts(file.GetDb().JsonDb); err != nil {
			logs.Error("hosts %s failed: %v", mode, err)
		} else {
			switch mode {
			case "backfill":
				logs.Info("hosts: migrated %d row(s) from hosts.json into SQLite", n)
			case "load":
				logs.Info("hosts: loaded %d row(s) from SQLite", n)
			}
		}
		// Phase 5: global — singleton Glob record.
		if mode, n, err := store.BackfillOrLoadGlobal(file.GetDb().JsonDb); err != nil {
			logs.Error("global %s failed: %v", mode, err)
		} else {
			switch mode {
			case "backfill":
				logs.Info("global: migrated %d row(s) from global.json into SQLite", n)
			case "load":
				logs.Info("global: loaded %d row(s) from SQLite", n)
			}
		}
		// Phase 6.1: app_settings — mirror nps.conf into SQLite. On
		// first boot we copy every non-bootstrap key from the in-memory
		// beego.AppConfig (just loaded from nps.conf); on subsequent
		// boots SQLite is the source of truth and we push values back
		// into beego.AppConfig so existing call sites read them
		// transparently.
		if mode, n, err := store.BackfillOrLoadAppSettings(file.GetDb().JsonDb); err != nil {
			logs.Error("app_settings %s failed: %v", mode, err)
		} else {
			switch mode {
			case "backfill":
				logs.Info("app_settings: migrated %d row(s) from nps.conf into SQLite", n)
			case "load":
				logs.Info("app_settings: applied %d row(s) from SQLite onto beego.AppConfig", n)
			}
		}
		// Phase 6.1: bridge.ServerTlsEnable was read at line 107 from
		// nps.conf only; refresh it now that SQLite has overridden the
		// in-memory beego.AppConfig (otherwise SQLite-stored tls_enable
		// is silently ignored on startup — listener never starts).
		bridge.ServerTlsEnable = beego.AppConfig.DefaultBool("tls_enable", false)
		registerSettingsHotHooks()
		// Wire the persistence hook so all subsequent Client CRUD
		// dual-writes through SQLite. Must happen AFTER backfill/load
		// so the bulk import does not recurse via UpsertClient.
		file.GetDb().JsonDb.ClientStore = store
		file.GetDb().JsonDb.TaskStore = store
		file.GetDb().JsonDb.HostStore = store
		file.GetDb().JsonDb.GlobalStore = store
	}
	task := &file.Tunnel{
		Mode: "webServer",
	}
	bridgePort, err := beego.AppConfig.Int("bridge_port")
	if err != nil {
		logs.Error("Getting bridge_port error", err)
		os.Exit(0)
	}

	logs.Info("日志路径：" + *npsLogPath)
	logs.Info("the config path is:" + common.GetRunPath())
	logs.Info("the version of server is %s ,allow client core version to be %s,tls enable is %t", version.VERSION, version.GetVersion(), bridge.ServerTlsEnable)
	connection.InitConnectionService()
	//crypt.InitTls(filepath.Join(common.GetRunPath(), "conf", "server.pem"), filepath.Join(common.GetRunPath(), "conf", "server.key"))
	crypt.InitTls()
	tool.InitAllowPort()
	tool.StartSystemInfo()
	timeout, err := beego.AppConfig.Int("disconnect_timeout")
	if err != nil {
		timeout = 60
	}
	go server.StartNewServer(bridgePort, task, beego.AppConfig.String("bridge_type"), timeout)
	// Adopt the implicit httpHostServer once StartNewServer publishes
	// it at RunList[0]; needed for HTTP/HTTPS proxy hot-restart.
	adoptInitialHttpHostServer()
}
