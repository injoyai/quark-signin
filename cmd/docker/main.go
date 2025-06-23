package main

import (
	"fmt"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/notice/pkg/push"
	"github.com/injoyai/notice/pkg/push/serverchan"
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
	vcode := cfg.GetString("vcode")
	_sign := cfg.GetString("sign")
	kps := cfg.GetString("kps")
	retry := cfg.GetInt("retry")
	spec := cfg.GetString("spec")
	sendKey := cfg.GetString("notice.serverChan.sendKey")

	logs.Info("版本:", "v1.2")
	logs.Info("说明:", "增加了通知失败的错误信息推送")
	logs.Info("==============================================================================")
	logs.Debug("Vcode:", vcode)
	logs.Debug("Sign:", _sign)
	logs.Debug("Kps:", kps)
	logs.Debug("Retry:", retry)
	logs.Debug("Spec:", spec)
	logs.Debug("SendKey:", sendKey)

	signin(vcode, _sign, kps, retry, sendKey)

	t := task.New()
	t.SetTask("signin", spec, func() { signin(vcode, _sign, kps, retry, sendKey) })
	t.Start()
	select {}
}
func signin(vcode, _sign, kps string, retry int, sendKey string) {
	s := &sign.Sign{
		Vcode: vcode,
		Sign:  _sign,
		Kps:   kps,
	}

	var info = new(sign.Info)
	var err error

	//判断是否签到
	for n := 0; n < retry; n++ {
		info, err = s.Info()
		if err != nil {
			logs.Err(err)
			<-time.After(time.Second * 10)
			continue
		}
		if info.Sign {
			//已经签到
			print(info, sendKey)
			return
		}
	}

	//进行签到
	if info == nil || !info.Sign {
		for x := 0; x < retry; x++ {
			if err = s.Signin(); err != nil {
				logs.Err(err)
				<-time.After(time.Second * 10)
				continue
			}
			info, err = s.Info()
			if err != nil {
				logs.Err(err)
				return
			}
			print(info, sendKey)
			break
		}
	}

	if err != nil {
		_notice(sendKey, fmt.Sprintf("签到失败, %s", err.Error()))
	}
}

func print(info *sign.Info, sendKey string) {
	msg := fmt.Sprintf("签到成功, 容量: %s, 进度:%d/%d, 总容量: %s", info.LastSpace, info.SignProgress, info.SignTarget, info.TotalSpace)
	logs.Info(msg)
	_notice(sendKey, msg)
}

func _notice(sendKey, msg string) {
	err := serverchan.New(sendKey).Push(&push.Message{
		Title:   "夸克签到",
		Content: msg,
	})
	logs.PrintErr(err)
}
