package sign

import (
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/oss"
)

type Sign struct {
	Vcode string `json:"vcode"`
	Sign  string `json:"sign"`
	Kps   string `json:"kps"`
}

func (this *Sign) do(method, u string) (*conv.Map, error) {
	if len(this.Vcode) == 0 {
		return nil, errors.New("未设置vcode")
	}
	if len(this.Sign) == 0 {
		return nil, errors.New("未设置sign")
	}
	if len(this.Kps) == 0 {
		return nil, errors.New("未设置kps")
	}
	resp := http.Url(u).SetQuerys(map[string]interface{}{
		"fr":    "android",
		"pr":    "ucpro",
		"vcode": this.Vcode,
		"sign":  this.Sign,
		"kps":   this.Kps,
	}).Debug().SetMethod(method).Do()
	if resp.Err() != nil {
		return nil, resp.Err()
	}
	return resp.GetBodyDMap().Get("data"), nil
}

// Info 获取签到信息
// {"status":200,"code":0,"message":"","timestamp":1736235143,"data":{"member_type":"NORMAL","use_capacity":29383673,"cap_growth":{"lost_total_cap":0,"cur_total_cap":440401920,"cur_total_sign_day":17},"88VIP":false,"member_status":{"Z_VIP":"UNPAID","VIP":"UNPAID","SUPER_VIP":"UNPAID","MINI_VIP":"UNPAID"},"cap_sign":{"sign_daily":true,"sign_target":7,"sign_daily_reward":41943040,"sign_progress":2,"sign_rewards":[{"name":"+20MB","reward_cap":20971520},{"name":"+40MB","highlight":"翻倍","reward_cap":41943040},{"name":"+20MB","reward_cap":20971520},{"name":"+20MB","reward_cap":20971520},{"name":"+20MB","reward_cap":20971520},{"name":"+20MB","reward_cap":20971520},{"name":"+100MB","highlight":"翻五倍","reward_cap":104857600}]},"cap_composition":{"other":0,"member_own":10737418240,"sign_reward":440401920},"total_capacity":11177820160},"metadata":{}}
func (this *Sign) Info() (*Info, error) {
	u := "https://drive-m.quark.cn/1/clouddrive/capacity/growth/info"
	m, err := this.do(http.MethodGet, u)
	if err != nil {
		return nil, err
	}
	return &Info{
		Vip:          m.GetBool("88VIP"),
		Sign:         m.GetBool("cap_sign.sign_daily"),
		SignNumber:   m.GetInt("cap_growth.cur_total_sign_day"),
		SignProgress: m.GetInt("cap_sign.sign_progress"),
		SignTarget:   m.GetInt("cap_sign.sign_target"),
		LastSpace:    oss.SizeString(m.GetInt64("cap_sign.sign_daily_reward")),
		TotalSpace:   oss.SizeString(m.GetInt64("cap_growth.cur_total_cap")),
		UseSpace:     oss.SizeString(m.GetInt64("use_capacity")),
	}, nil
}

// Signin 签到
func (this *Sign) Signin() error {
	u := "https://drive-m.quark.cn/1/clouddrive/capacity/growth/sign"
	_, err := this.do(http.MethodPost, u)
	return err
}

type Info struct {
	Vip          bool   //是否是88vip
	Sign         bool   //是否签到
	SignNumber   int    //签到次数
	SignProgress int    //签到进度,1~7
	SignTarget   int    //签到目标,7
	LastSpace    string //本次签到空间
	TotalSpace   string //总签到空间
	UseSpace     string //已使用空间
}
