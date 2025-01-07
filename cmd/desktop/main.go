package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/quark-signin"
	"github.com/injoyai/tool/config"
	"time"
)

func main() {
	filename := oss.UserInjoyDir("/quark-signin/config.yaml")

	oss.NewNotExist(filename, `
cookie:
    kps: 
    sign: 
    vcode: 
sys:
    retry: "3"
    startup-sign: false
    timer-corn: 0 0 9 * * *
    timer-sign: true
`)

	file := cfg.WithFile(filename).(*conv.Map)
	cfg.Init(file)

	var t *task.Cron

	setHint := func(info *sign.Info, err error) {}

	f := func(m *conv.Map) {
		s := &sign.Sign{
			Vcode: m.GetString("cookie.vcode"),
			Sign:  m.GetString("cookie.sign"),
			Kps:   m.GetString("cookie.kps"),
		}
		retry := m.GetInt("sys.retry", 3)
		for n := 0; n < retry; n++ {
			info, err := s.Info()
			if err != nil {
				logs.Err(err)
				<-time.After(time.Second)
				continue
			}
			for x := 0; x < retry; x++ {
				if !info.Sign {
					if err := s.Signin(); err != nil {
						logs.Err(err)
						<-time.After(time.Second)
						continue
					}
					return
				}
			}
		}
	}

	//开机签到
	if cfg.GetBool("sys.startup-sign") {
		logs.Debug("开机签到")
		f(file)
	}

	tray.Run(
		func(s *tray.Tray) {
			s.SetIco(Ico)
			setHint = func(info *sign.Info, err error) {
				format := "签到状态: %v\n签到进度: %d/%d\n签到空间: %s\n错误消息: %v"
				if err != nil {
					s.SetHint(fmt.Sprintf(format, "未知", 0, 0, "", err.Error()))
					return
				}
				s.SetHint(fmt.Sprintf(format, info.Sign, info.SignProgress, info.SignTarget, info.LastSpace, "无"))
			}
			//设置提示
			x := &sign.Sign{
				Vcode: cfg.GetString("cookie.vcode"),
				Sign:  cfg.GetString("cookie.sign"),
				Kps:   cfg.GetString("cookie.kps"),
			}
			info, err := x.Info()
			logs.PrintErr(err)
			setHint(info, err)
		},
		tray.WithShow(func(m *tray.Menu) {
			config.GUI(
				&config.Config{
					Width:    720,
					Height:   720,
					Filename: filename,
					Natures: config.Natures{
						{Name: "系统", Key: "sys", Type: "object2", Value: config.Natures{
							{Name: "重试次数", Key: "retry", Type: "int"},
							{Name: "开机签到", Key: "startup-sign", Type: "bool"},
							{Name: "定时签到", Key: "timer-sign", Type: "bool"},
							{Name: "定时Corn", Key: "timer-corn"},
						}},
						{Name: "Cookie", Key: "cookie", Type: "object2", Value: config.Natures{
							{Name: "vcode", Key: "vcode"},
							{Name: "sign", Key: "sign"},
							{Name: "kps", Key: "kps"},
						}},
					},
					OnSaved: func(m *conv.Map) {
						//定时签到
						if m.GetBool("sys.timer-sign") {
							if t != nil {
								t.Stop()
							}
							t = task.New().Start()
							t.SetTask("", m.GetString("sys.timer-corn"), func() {
								logs.Debug("定时签到")
								f(m)
							})
						}

					},
				},
			)

		}, tray.Name("配置"), tray.Icon(tray.IconSetting)),
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
	)

}
