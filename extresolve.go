package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/InVisionApp/tabular"
	"github.com/swaggo/cli"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var app = cli.NewApp()
var tab tabular.Table

type DnsRecord struct {
	IP          string  `json:"ip"`
	Name        string  `json:"name"`
	City        string  `json:"city"`
	Dnssec      bool    `json:"dnssec"`
	Reliability float64 `json:"reliability"`
}

func chechHost(c *cli.Context) (err error) {
	count := c.Int("count")

	protocol := c.String("protocol")

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
	f := func(bool) string {
		if detail {
			return tab.Print("DNS", "Name", "City", "Dnssec", "Reliability", "Result")
		} else {
			return tab.Print("DNS", "Name", "City", "Result")
		}
	}
	format := f(detail)

	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		var dnsHosts []DnsRecord
		if err = json.Unmarshal(body, &dnsHosts); err == nil {
			for id, record := range dnsHosts {
				var resolver *net.Resolver
				resolver = &net.Resolver{
					PreferGo: true,
					Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
						d := net.Dialer{
							Timeout: time.Duration(timeout) * time.Second,
						}
						return d.DialContext(ctx, protocol, net.JoinHostPort(record.IP, "53"))
					},
				}
				ips, _ := resolver.LookupIPAddr(context.Background(), host)
				var resolveHost = ""
				for _, ip := range ips {
					resolveHost = resolveHost + " " + ip.String()
				}
				if detail {
					fmt.Printf(format, record.IP, record.Name, record.City, record.Dnssec, record.Reliability, resolveHost)
				} else {
					fmt.Printf(format, record.IP, record.Name, record.City, resolveHost)
				}

				if (id + 1) == count {
					break
				}
			}
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
			Name:  "protocol",
			Value: "udp",
			Usage: "connect protocol (udp or tcp)",
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
	// Tabular settings
	tab = tabular.New()
	tab.Col("DNS", "DNS server", 20)
	tab.Col("Name", "DNS name", 60)
	tab.Col("City", "DNS city", 20)
	tab.Col("Dnssec", "DNS security", 12)
	tab.Col("Reliability", "DNS reliability", 15)
	tab.Col("Result", "DNS response", 0)
	//
	app.Action = func(c *cli.Context) {
		if err := chechHost(c); err != nil {
			log.Fatal(err)
		}
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
