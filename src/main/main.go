package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	id     = os.Getenv("id")
	key    = os.Getenv("id")
	bucket = os.Getenv("id")
	region = os.Getenv("id")
)

var (
	info *Info = &Info{}
)

type Info struct {
	cc *cos.Client
}

func initCos() {
	//将<bucket>和<region>修改为真实的信息
	//bucket的命名规则为{name}-{appid} ，此处填写的存储桶名称必须为此格式
	u, _ := url.Parse("https://" + bucket + ".cos." + region + ".myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		//设置超时时间
		Timeout: 100 * time.Second,
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  id,
			SecretKey: key,
		},
	})
	info.cc = c
}

func DescribeBucketList() []cos.Object {
	list, resp, err := info.cc.Bucket.Get(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("%s\n", string(bs))
	fmt.Println("\n\n")
	for _, v := range list.Contents {
		fmt.Println(v.Key)
	}
	if list == nil && len(list.Contents) == 0 {
		return nil
	}
	return list.Contents
}

func DelBucketList(object []cos.Object) error {
	_, _, err := info.cc.Object.DeleteMulti(context.Background(), &cos.ObjectDeleteMultiOptions{
		Objects: object,
	})
	return err
}

func UploadMultiObject() {

}
