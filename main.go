package main

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/colorstring"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"strings"
)

type NodeService struct {
	Service string
	Address string
	Port    int
	Tags    []string
}

type ConsulNode struct {
	Node     string
	Services []NodeService
}

func main() {
	app := cli.NewApp()
	app.Usage = "Decode micro registry's tag of metadata and endpoint"
	app.Action = dec
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "data, d",
			Usage: "encoded string",
		},
		cli.StringFlag{
			Name:  "url, u",
			Usage: "consul address",
			Value: "http://127.0.0.1:8500/v1/internal/ui/nodes?dc=dc1&token=",
		},
		cli.StringSliceFlag{
			Name:  "service, s",
			Usage: "services you wanted to decode",
		},
		cli.StringSliceFlag{
			Name:  "keyword, k",
			Usage: "highlight keywords",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		return
	}
}

func dec(c *cli.Context) (err error) {

	data := c.String("data")
	keywords := c.StringSlice("keyword")

	if len(data) > 0 {
		return decData(data, keywords)
	}

	url := c.String("url")
	if len(url) == 0 {
		return
	}

	serviceList := c.StringSlice("service")
	var srvMap map[string]bool
	if len(serviceList) > 0 {
		srvMap = make(map[string]bool)
		for i := 0; i < len(serviceList); i++ {
			srvMap[serviceList[i]] = true
		}
	}

	var resp *http.Response
	if resp, err = http.DefaultClient.Get(url); err != nil {
		return
	}

	defer resp.Body.Close()

	nodes := []ConsulNode{}

	decoder := json.NewDecoder(resp.Body)

	if err = decoder.Decode(&nodes); err != nil {
		return
	}

	for i := 0; i < len(nodes); i++ {
		for j := 0; j < len(nodes[i].Services); j++ {
			service := nodes[i].Services[j]
			if srvMap != nil {
				if _, exist := srvMap[service.Service]; !exist {
					continue
				}
			}

			metadata, endpoints := decodeTags(service.Tags)
			dataMap := map[string]interface{}{
				"service":   service.Service,
				"address":   service.Address,
				"port":      service.Port,
				"metadata":  metadata,
				"endpoints": endpoints,
			}

			var out []byte
			if out, err = json.MarshalIndent(dataMap, "", "    "); err != nil {
				return
			}

			strJSON := string(out)
			for i := 0; i < len(keywords); i++ {
				strJSON = strings.Replace(strJSON, keywords[i], "[yellow]"+keywords[i]+"[white]", -1)
			}

			colorstring.Println(strJSON)
		}
	}
	return
}

func decData(data string, keywords []string) (err error) {
	data = strings.TrimSpace(data)

	tags := strings.Split(data, ",")
	metadata, endpoints := decodeTags(tags)

	dataMap := map[string]interface{}{"metadata": metadata, "endpoints": endpoints}

	var out []byte
	if out, err = json.MarshalIndent(dataMap, "", "    "); err != nil {
		return
	}

	strJSON := string(out)
	for i := 0; i < len(keywords); i++ {
		strJSON = strings.Replace(strJSON, keywords[i], "[yellow]"+keywords[i]+"[white]", -1)
	}

	colorstring.Println(strJSON)

	return
}

func decodeTags(tags []string) (metadata interface{}, endpoints interface{}) {

	var metadataTags []string
	var endpointTags []string

	for i := 0; i < len(tags); i++ {
		tag := strings.TrimSpace(tags[i])
		if strings.HasPrefix(tag, "t") {
			metadataTags = append(metadataTags, tag)
		} else if strings.HasPrefix(tag, "e") {
			endpointTags = append(endpointTags, tag)
		}
	}

	metadata = decodeMetadata(metadataTags)
	endpoints = decodeEndpoints(endpointTags)
	return
}
