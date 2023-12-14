package aliyun

import (
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/pkg6/go-sms"
	"time"
)

//https://help.aliyun.com/document_detail/419273.html

type ALiYun struct {
	Host            string `json:"host" xml:"host"`
	RegionId        string `json:"region_id" xml:"region_id"`
	AccessKeyId     string `json:"access_key_id" xml:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" xml:"access_key_secret"`
	Format          string `json:"format" xml:"format"`
	gosms.Lock
}

func GateWay(accessKeyId, accessKeySecret string) gosms.IGateway {
	gateway := &ALiYun{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	return gateway.I()
}
func (g ALiYun) I() gosms.IGateway {
	if g.Host == "" {
		g.Host = "http://dysmsapi.aliyuncs.com"
	}
	if g.RegionId == "" {
		g.RegionId = "cn-hangzhou"
	}
	if g.Format == "" {
		g.Format = "JSON"
	}
	g.LockInit()
	return &g
}

func (g *ALiYun) AsName() string {
	return "aliyun"
}

// 请求参数生成
func (g *ALiYun) query() gosms.MapStrings {
	if g.RegionId == "" {
		g.RegionId = "cn-hangzhou"
	}
	if g.Format == "" {
		g.Format = "JSON"
	}
	maps := gosms.MapStrings{
		"AccessKeyId":      g.AccessKeyId,
		"Action":           "SendSms",
		"Format":           g.Format,
		"RegionId":         g.RegionId,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   gosms.Uniqid("gosms"),
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}
	return maps
}

func (g *ALiYun) Send(to gosms.IPhoneNumber, message gosms.IMessage) (gosms.SMSResult, error) {
	g.Lock.L.Lock()
	defer g.L.Unlock()
	data := message.GetData(g.I())
	data.Delete("signName")

	client, err := dysmsapi.NewClientWithAccessKey(g.RegionId, g.AccessKeyId, g.AccessKeySecret)

	request := dysmsapi.CreateSendSmsRequest()                                 //创建请求
	request.Scheme = "https"                                                   //请求协议
	request.PhoneNumbers = gosms.GetPhoneNumber(to)                            //接收短信的手机号码
	request.SignName = data.GetDefault("signName", message.GetSignName(g.I())) //短信签名名称
	request.TemplateCode = message.GetTemplate(g.I())                          //短信模板ID
	par, _ := data.ToJson()
	request.TemplateParam = par //将短信模板参数传入短信模板

	response, err := client.SendSms(request) //调用阿里云API发送信息
	if err != nil {                          //处理错误
		return gosms.SMSResult{}, err
	} else {
		if response.Code != "OK" {
			return gosms.SMSResult{}, errors.New(response.Message)
		} else {
			return gosms.SMSResult{}, nil
		}
	}
}
