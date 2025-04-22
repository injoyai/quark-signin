package main

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	sign "github.com/injoyai/quark-signin"
	"time"
)

func init() {
	cfg.Init(
		cfg.WithFile("./config/config.yaml"),
		cfg.WithEnv(),
	)
}

func main() {
	signin()
	spec := cfg.GetString("spec")
	t := task.New()
	t.SetTask("signin", spec, signin)
	t.Start()
	select {}
}
func signin() {
	vcode := cfg.GetString("vcode")
	_sign := cfg.GetString("sign")
	kps := cfg.GetString("kps")
	retry := cfg.GetInt("retry")

	s := &sign.Sign{
		Vcode: vcode,
		Sign:  _sign,
		Kps:   kps,
	}
	for n := 0; n < retry; n++ {
		info, err := s.Info()
		if err != nil {
			logs.Err(err)
			<-time.After(time.Second * 10)
			continue
		}
		if info.Sign {
			//已经签到
			print(info, nil)
			return
		}
		for x := 0; x < retry; x++ {
			if !info.Sign {
				if err := s.Signin(); err != nil {
					logs.Err(err)
					<-time.After(time.Second * 10)
					continue
				}
				info, err := s.Info()
				if err != nil {
					logs.Err(err)
					return
				}
				print(info, nil)
				return
			}
		}
	}
}

func print(info *sign.Info, err error) {
	if err != nil {
		logs.Err(err)
		return
	}
	logs.Infof("状态: %v, 容量: %s, 进度:%d/%d, 总容量: %s", info.Sign, info.LastSpace, info.SignProgress, info.SignTarget, info.TotalSpace)
}
