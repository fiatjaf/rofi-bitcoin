package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/fiatjaf/go-lnurl"
	"github.com/imroc/req"
)

func main() {
	v := strings.TrimSpace(strings.Join(os.Args[1:], " "))

	if v == "" {
		fmt.Printf("\x00prompt\x1fbitcoin\n")

		w1 := make(chan bool, 1)
		go func() {
			if resp, err := req.Get("https://ln.bigsun.xyz/api/nodes?select=pubkey,alias,capacity&capacity=gt.100000&alias=neq.&order=capacity.desc"); err == nil {
				var lnnodes []LNNode
				if err := resp.ToJSON(&lnnodes); err == nil {
					for _, n := range lnnodes {
						fmt.Printf("ln: %dk %s [%s]\n", n.Capacity/1000, n.Alias, n.PubKey[:7])
					}
				}
			}
			w1 <- true
		}()

		w2 := make(chan bool, 1)
		go func() {
			if resp, err := req.Get("https://blockstream.info/api/blocks/tip/height"); err == nil {
				if tip, err := strconv.ParseInt(resp.String(), 64, 10); err == nil {
					for i := tip; i > tip-5; i-- {
						fmt.Printf("block: %d\n", i)
					}
				}
			}
			w2 <- true
		}()

		<-w1
		<-w2
	} else {
		var target string

		switch strings.Split(v, " ")[0] {
		case "ln:":
			m := lnNodeRe.FindStringSubmatch(v)[1]
			target = "https://ln.bigsun.xyz/" + m
		case "block:":
			b := strings.Split(v, ":")[1]
			target = "https://blockstream.info/" + b
		default:
			lowerV := strings.ToLower(v)
			if strings.HasPrefix(lowerV, "bitcoin:") {
				lowerV = lowerV[8:]
			}
			if strings.HasPrefix(lowerV, "lightning:") {
				lowerV = lowerV[10:]
			}

			switch {
			case strings.HasPrefix(lowerV, "lnbc") ||
				strings.HasPrefix(lowerV, "lntb") ||
				strings.HasPrefix(lowerV, "lnbcrt"):
				target = "https://lndecode.com/?invoice=" + target
			case strings.HasPrefix(lowerV, "lnurl1"):
				decoded, err := lnurl.LNURLDecode(lowerV)
				if err != nil {
					fmt.Printf("\x00message\x1f<b>lnurl</b>: " + err.Error() + "\n")
					return
				}
				target = decoded
			case shortChannelIdRe.MatchString(lowerV):
				target = "https://ln.bigsun.xyz/" + lowerV
			default:
				target = "https://blockstream.info/" + v
			}
		}

		exec.Command("xdg-open", target).Run()
	}
}

var lnNodeRe = regexp.MustCompile(`\[([a-e0-9]+)\]`)
var shortChannelIdRe = regexp.MustCompile(`^\d+x\d+x\d+$`)

type LNNode struct {
	PubKey   string `json:"pubkey"`
	Capacity int64  `json:"capacity"`
	Alias    string `json:"alias"`
}
