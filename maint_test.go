package main

import (
	"gopkg.in/h2non/gock.v1"
	"os"
	"testing"
)

//func TestDiscovery_Run(t *testing.T) {
//	jsonFile, err := os.Open("vips-test.json")
//	// if we os.Open returns an error then handle it
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println("Successfully Opened vips-test.json")
//	// defer the closing of our jsonFile so that we can parse it later on
//	defer jsonFile.Close()
//
//	var nodes []*L4LBService
//	dec := json.NewDecoder(jsonFile)
//	err = dec.Decode(&nodes)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println("Successfully parsed vips-test.json")
//
//	cfg := sdConfig{
//		TagSeparator:    ",",
//		Address:         "127.0.0.1:62080",
//		RefreshInterval: 30,
//		Prefix:          "prometheus",
//		Framework:       "*",
//	}
//
//	disc, err := newDiscovery(cfg)
//	var tgs []*targetgroup.Group
//	for _, service := range nodes {
//		if service.Protocol == "tcp" && strings.HasPrefix(service.Vip, disc.prefix) {
//			tg, err := disc.parseServiceNodes(service)
//			if err != nil {
//				fmt.Printf("error parsing services nodes %s\n", err.Error())
//				continue
//			}
//			tgs = append(tgs, tg)
//		}
//	}
//
//	for index, tg := range tgs {
//		fmt.Printf("Target Group index: %d\n", index)
//		fmt.Printf("Lables: %s\n", tg.Labels.String())
//		fmt.Printf("Targets: %s\n", tg.Targets)
//	}
//}

func TestFoo(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution

	gock.New("http://localhost:62080").
		Get("/v1/vips").
		Reply(200).
		File("vips-test.json").
		SetHeader("Content-Type", "application/json")
		//JSON(map[string]string{"foo": "bar"})


	// Your test code starts here...
	os.Args = []string{"prometheus-sd-dcosl4lb-linux", "--output.file=/root/output.json" }
	//
	main()
}