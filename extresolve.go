package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/swaggo/cli"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var app = cli.NewApp()

type DnsRecord struct {
	IP          string  `json:"ip"`
	Name        string  `json:"name"`
	City        string  `json:"city"`
	Dnssec      bool    `json:"dnssec"`
	Reliability float64 `json:"reliability"`
}

type Header *simpletable.Header

func chechHost(c *cli.Context) (err error) {
	// Tab
	tab := simpletable.New()

	count := c.Int("count")

	host := c.String("host")

	geo := c.String("region")

	timeout := c.Int("timeout")

	timeoutGetDns := c.Int("timeout-get-dns-list")

	detail := c.Bool("detail")

	var httpClient = &http.Client{
		Timeout: time.Duration(timeoutGetDns) * time.Second,
	}
	resp, err := httpClient.Get("https://public-dns.info/nameserver/" + geo + ".json")
	if err != nil {
		return error(err)
	}

	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		var dnsHosts []DnsRecord
		if err = json.Unmarshal(body, &dnsHosts); err == nil {
			for id, record := range dnsHosts {
				var resolver *net.Resolver

				ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				ips, _ := resolver.LookupIPAddr(ctx, host)

				var resolveHost = ""
				for _, ip := range ips {
					resolveHost = resolveHost + " " + ip.String()
				}
				if detail {
					tab.Header = &simpletable.Header{
						Cells: []*simpletable.Cell{
							{Align: simpletable.AlignCenter, Text: "DNS"},
							{Align: simpletable.AlignCenter, Text: "Name"},
							{Align: simpletable.AlignCenter, Text: "City"},
							{Align: simpletable.AlignCenter, Text: "Dnssec"},
							{Align: simpletable.AlignCenter, Text: "Reliability"},
							{Align: simpletable.AlignCenter, Text: "Result"},
						},
					}
					r := []*simpletable.Cell{
						{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", record.IP)},
						{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%s", record.Name)},
						{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%s", record.City)},
						{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%t", record.Dnssec)},
						{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%f", record.Reliability)},
						{Align: simpletable.AlignRight, Text: fmt.Sprintf("%s", resolveHost)},
					}

					tab.Body.Cells = append(tab.Body.Cells, r)
				} else {
					tab.Header = &simpletable.Header{
						Cells: []*simpletable.Cell{
							{Align: simpletable.AlignCenter, Text: "DNS"},
							{Align: simpletable.AlignCenter, Text: "Name"},
							{Align: simpletable.AlignCenter, Text: "City"},
							{Align: simpletable.AlignCenter, Text: "Result"},
						},
					}
					r := []*simpletable.Cell{
						{Align: simpletable.AlignRight, Text: fmt.Sprintf("%s", record.IP)},
						{Align: simpletable.AlignRight, Text: fmt.Sprintf("%s", record.Name)},
						{Align: simpletable.AlignRight, Text: fmt.Sprintf("%s", record.City)},
						{Align: simpletable.AlignRight, Text: fmt.Sprintf("%s", resolveHost)},
					}

					tab.Body.Cells = append(tab.Body.Cells, r)
				}

				if (id + 1) == count {
					break
				}
			}
			tab.SetStyle(simpletable.StyleCompactLite)
			fmt.Println(tab.String())
		} else {
			return error(err)
		}
	} else {
		return error(err)
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	// Cli settings
	app.Name = "Dns check"
	app.Version = "1.0"
	app.Usage = "Checking current host in all public dns in current geo area.\r\n" +
		"\t Uses public-dns.info for getting dns servers in selected geo area."
	app.UsageText = "check_dns -n {host} -r {geo_area}"
	app.HideHelp = false
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "Dmitriy Vlassov",
			Email: "dmitriy@vlassov.kz",
		},
	}
	app.Flags = []cli.Flag{
		cli.Int64Flag{
			Name:  "count, c",
			Value: 10,
			Usage: "quantity checks",
		},
		cli.StringFlag{
			Name:     "host, n",
			Value:    "",
			Required: true,
			Usage:    "host for checks",
		},
		cli.StringFlag{
			Name:     "region, r",
			Value:    "",
			Required: true,
			Usage:    "region for checks",
		},
		cli.Int64Flag{
			Name:  "timeout, t",
			Value: 3,
			Usage: "timeout for check",
		},
		cli.Int64Flag{
			Name:  "timeout-get-dns-list",
			Value: 30,
			Usage: "timeout for get dns from public-dns.info service",
		},
		cli.BoolFlag{
			Name:  "detail, d",
			Usage: "more info about dns servers",
		},
	}
	app.Action = func(c *cli.Context) {
		if err := chechHost(c); err != nil {
			log.Fatal(err)
		}
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
