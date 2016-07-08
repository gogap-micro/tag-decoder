package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Usage = "Decode micro registry's tag of metadata and endpoint"
	app.Action = dec
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "data, d",
			Usage: "Encoded string",
		},
	}

	if err := app.Run(os.Args); err != nil {
		return
	}
}

func dec(c *cli.Context) (err error) {

	data := c.String("data")

	if len(data) == 0 {
		return
	}

	data = strings.TrimSpace(data)
	tags := strings.Split(data, ",")

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

	metadata := decodeMetadata(metadataTags)
	endpoints := decodeEndpoints(endpointTags)

	dataMap := map[string]interface{}{"metadata": metadata, "endpoints": endpoints}

	var out []byte
	if out, err = json.MarshalIndent(dataMap, "", "    "); err != nil {
		return
	}

	fmt.Println(string(out))

	return
}
