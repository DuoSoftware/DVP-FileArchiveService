package main

import (
	"github.com/asaskevich/govalidator"
	"github.com/robfig/cron"
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"os"
	//"path/filepath"
	"io/ioutil"
	"path/filepath"
)

type Response struct{
	Exception string
	CustomMessage string
	IsSuccess bool
	Result []Item
}


type Item struct {
	CompanyId  int
	SpaceLimit Limit
}


type Limit struct{
	SpaceLimit int
	SpaceType  string
	SpaceUnit  string
}

type Config struct {
	RootPath string
	Services struct {
			 AccessToken        string
			 UserServiceHost    string
			 UserServicePort    string
			 UserServiceVersion string
		 }
}


func loadConfig() Config{

	dirPath := GetDirPath()
	confPath := filepath.Join(dirPath, "default.json")
	fmt.Println("GetDefaultConfig config path: ", confPath)
	content, operr := ioutil.ReadFile(confPath)
	if operr != nil {
		fmt.Println(operr)
	}

	defConfiguration := Config{}
	json.Unmarshal(content, &defConfiguration)

	////////////////////////load envs/////////////////////////////////////
	envConfPath := filepath.Join(dirPath, "custom-environment-variables.json")

	envContent, operr := ioutil.ReadFile(envConfPath)


	if operr != nil {
		fmt.Println(operr)
	}else {

		defEnvConfiguration := Config{}
		unErr := json.Unmarshal(envContent, &defEnvConfiguration)

		fmt.Println(defConfiguration)

		if unErr != nil {

			if defEnvConfiguration.Services.AccessToken != "" {

				defConfiguration.Services.AccessToken = os.Getenv(defEnvConfiguration.Services.AccessToken)
			}

			if defEnvConfiguration.Services.UserServiceHost != "" {

				defConfiguration.Services.UserServiceHost = os.Getenv(defEnvConfiguration.Services.UserServiceHost)
			}

			if defEnvConfiguration.Services.UserServiceVersion != "" {

				defConfiguration.Services.UserServiceVersion = os.Getenv(defEnvConfiguration.Services.UserServiceVersion)
			}

		}
	}


	return defConfiguration
}

func checkFile(path string, info os.FileInfo, err error) error{
	fmt.Println(path,info,err)
	fmt.Println(info.ModTime())
	return nil
}

func main() {



	//fmt.Println(resp, err)

	c := cron.New()
	fmt.Println("@every 10s")
	//@midnight
	//@every 10s

	conf := loadConfig()

	rootPath := conf.RootPath
	accessToken := conf.Services.AccessToken
	//"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJzdWtpdGhhIiwianRpIjoiYWEzOGRmZWYtNDFhOC00MWUyLTgwMzktOTJjZTY0YjM4ZDFmIiwic3ViIjoiNTZhOWU3NTlmYjA3MTkwN2EwMDAwMDAxMjVkOWU4MGI1YzdjNGY5ODQ2NmY5MjExNzk2ZWJmNDMiLCJleHAiOjE5MDIzODExMTgsInRlbmFudCI6LTEsImNvbXBhbnkiOi0xLCJzY29wZSI6W3sicmVzb3VyY2UiOiJhbGwiLCJhY3Rpb25zIjoiYWxsIn1dLCJpYXQiOjE0NzAzODExMTh9.Gmlu00Uj66Fzts-w6qEwNUz46XYGzE8wHUhAJOFtiRo"
	authToken := fmt.Sprintf("Bearer %s", accessToken)



	url := fmt.Sprintf("http://%s:%s/DVP/API/%s/Organisation/SpaceLimits/VoiceClip",conf.Services.UserServiceHost, conf.Services.UserServicePort, conf.Services.UserServiceVersion)

	isValid := govalidator.IsIP(conf.Services.UserServiceHost)
	if isValid != true{
		url = fmt.Sprintf("http://%s/DVP/API/%s/Organisation/SpaceLimits/VoiceClip",conf.Services.UserServiceHost, conf.Services.UserServiceVersion)
	}

	fmt.Println(url)
	fmt.Println("Every midnight")
	crErr := c.AddFunc("@every 10s", func() {


		fmt.Println("@every 10s")
		req, err := http.NewRequest("GET", url , nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("authorization", authToken)
		req.Header.Set("companyinfo", "1:1")

		client := &http.Client{}
		res := Response{}
		res.Result = make([]Item, 0)

		resp, err := client.Do(req)

		fmt.Println(resp, err)
		if err == nil && resp.StatusCode == 200 {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			json.Unmarshal(bodyBytes, &res)

			for _, item := range res.Result {

				go func(itm Item) {

					companyDirectory := fmt.Sprintf("Company_%d_Tenant_1", item.CompanyId)
					companyPath := filepath.Join(rootPath, companyDirectory)
					conversationPath := filepath.Join(companyPath, "CONVERSATION")
					fmt.Println(conversationPath)

					files, _ := ioutil.ReadDir(conversationPath)
					for _, info := range files {
						fmt.Println(info.Name())

						t, err := time.Parse("2006-1-2", info.Name())
						fmt.Println(t, err)
						if err == nil {

							fmt.Println(t.Before(time.Now().AddDate(0, 0, -(itm.SpaceLimit.SpaceLimit + 1))))
							if (t.Before(time.Now().AddDate(0, 0, -(itm.SpaceLimit.SpaceLimit + 1)))) {

								if info.IsDir() {
									fileToDelete := filepath.Join(conversationPath,info.Name());
									fmt.Println(fileToDelete)
									os.RemoveAll(fileToDelete)

								}
							}
						}
					}
				}(item)

			}

			//fmt.Println(res.Result)
		}

		defer resp.Body.Close()
	})

	fmt.Println(crErr)
	c.Start()

	for ; ; {
		time.Sleep(1 * time.Second)
	}

}


func GetDirPath() string {
	envPath := os.Getenv("GO_CONFIG_DIR")
	if envPath == "" {
		envPath = "D:\\PROJECTS\\VEERY\\DVP-FileArchiveService"
	}
	fmt.Println(envPath)

	envPath = filepath.Join(envPath,"config")

	return envPath
}


