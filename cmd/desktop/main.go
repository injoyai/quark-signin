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
	cfg.Init(cfg.WithFile(filename))

	var t *task.Cron

	setHint := func(info *sign.Info, err error) {}

	f := func(m *conv.Map, startup bool) {
		s := &sign.Sign{
			Vcode: m.GetString("cookie.vcode"),
			Sign:  m.GetString("cookie.sign"),
			Kps:   m.GetString("cookie.kps"),
		}

		do := func() {
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

						break
					}
				}
			}
		}

		//开机自启
		if startup {
			do()
			return
		}

		//定时器
		if m.GetBool("sys.timer-sign") {
			if t != nil {
				t.Stop()
			}
			t = task.New().Start()
			t.SetTask("", m.GetString("sys.timer-corn"), do)
		}

	}

	//开机签到
	f(cfg.GetDMap(""), cfg.GetBool("sys.startup-sign"))

	tray.Run(
		func(s *tray.Tray) {
			setHint = func(info *sign.Info, err error) {
				format := "签到状态: %v\n签到进度: %d/%d\n获取空间: %s"
				if err != nil {
					s.SetHint(fmt.Sprintf(format, err.Error(), 0, 0, ""))
					return
				}
				s.SetHint(fmt.Sprintf(format, info.Sign, info.SignProgress, info.SignTarget, info.LastSpace))
			}
			x := &sign.Sign{
				Vcode: cfg.GetString("cookie.vcode"),
				Sign:  cfg.GetString("cookie.sign"),
				Kps:   cfg.GetString("cookie.kps"),
			}
			info, err := x.Info()
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
					OnSaved: func(m *conv.Map) { f(m, false) },
				},
			)

		}, tray.Name("配置"), tray.Icon(tray.IconSetting)),
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
	)

}
