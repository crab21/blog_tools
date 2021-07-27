package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	id         *string
	key        *string
	bucket     *string
	region     *string
	excludeDir *string
	path       *string
)

var (
	info          *Info = &Info{}
	excludeDirMap       = map[string]interface{}{}
)

type Info struct {
	cc *cos.Client
}

func main() {
	path = flag.String("Path", "", "start path")
	id = flag.String("SecretID", "", "ID")
	key = flag.String("SecretKey", "", "ID")
	bucket = flag.String("BucketName", "", "bucket")
	region = flag.String("Region", "", "region")
	excludeDir = flag.String("ExcludeFilePaths", "", "ExcludeFilePaths:   a:b:c")
	flag.Parse()
	if *id == "" || *key == "" || *bucket == "" || *region == "" || *path == "" {
		err := fmt.Errorf("params error : ID/Key/BucketName/Region is not emtpy")
		fmt.Println(err.Error(), "--------->error")
		return
	}

	initCos()
	excludeDirInit()
	lls := DescribeBucketList()
	err := DelBucketList(lls)
	fmt.Println(err, "--------->error")
	s := *path
	result := GetDirFileNameAndPath(map[string]string{}, s)
	UploadMultiObject(result)
}

func initCos() {
	//将<bucket>和<region>修改为真实的信息
	//bucket的命名规则为{name}-{appid} ，此处填写的存储桶名称必须为此格式
	u, _ := url.Parse("https://" + *bucket + ".cos." + *region + ".myqcloud.com")
	fmt.Println(u, "--------->error")
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		//设置超时时间
		Timeout: 30 * time.Minute,
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  *id,
			SecretKey: *key,
		},
	})
	info.cc = c
}

func DescribeBucketList() []cos.Object {
	list, _, err := info.cc.Bucket.Get(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	// bs, _ := ioutil.ReadAll(resp.Body)
	// resp.Body.Close()
	// fmt.Printf("%s\n", string(bs))
	// fmt.Println("\n\n")

	if list == nil && len(list.Contents) == 0 {
		return nil
	}
	return list.Contents
}

func DelBucketList(object []cos.Object) error {
	for _, v := range object {
		if !excludeDirPass(v.Key) {
			continue
		}

		fmt.Println("del cos file:  ", v.Key)

		_, err := info.cc.Object.Delete(context.Background(), v.Key, nil)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func excludeDirInit() {
	if *excludeDir == "" {
		return
	}
	s := strings.Split(*excludeDir, ":")
	for _, v := range s {
		excludeDirMap[v] = nil
	}
	fmt.Println(excludeDirMap, ".................")
}

func excludeDirPass(path string) bool {
	if len(excludeDirMap) == 0 {
		return true
	}
	for K, _ := range excludeDirMap {
		if K == "" {
			continue
		}

		if strings.Contains(path, K) {
			return false
		}
	}
	return true

}

func UploadMultiObject(result map[string]string) {
	for v, _ := range result {
		if !excludeDirPass(v) {
			continue
		}
		fmt.Println("upload cos file:  ", v)
		info.cc.Object.PutFromFile(context.Background(), v, v, &cos.ObjectPutOptions{})
	}

}

func GetFileNameAndPath() (result map[string]string) {
	result = make(map[string]string, 0)

	return
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsFile(path string) bool {
	return !IsDir(path)
}

func GetDirFileNameAndPath(result map[string]string, path string) map[string]string {
	if !Exists(path) {
		return result
	}

	if !IsDir(path) {
		return result
	}

	file, _ := ioutil.ReadDir(path)
	for _, v := range file {
		if v.IsDir() {
			result = GetDirFileNameAndPath(result, path+"/"+v.Name())
			continue
		}
		result[path+"/"+v.Name()] = v.Name()
		// fmt.Println("upload file:  ", path+"/"+v.Name())
	}
	return result
}
