package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/m-lab/tcp-info/collector"

	"github.com/m-lab/go/flagx"
)

var (
	reps    = flag.Int("reps", 0, "How many cycles should be recorded, 0 means continuous")
	outfile = flag.String("file", "", "File to write to.")
)

func main() {
	flag.Parse()
	flagx.ArgsFromEnv(flag.CommandLine)

	if !strings.HasSuffix(*outfile, ".jsonl") {
		log.Fatal("must specify -file ending in .jsonl")
	}

	o, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer o.Close()

	for i := 0; *reps == 0 || i < *reps; i++ {
		res6, err := collector.OneType(syscall.AF_INET6)
		if err != nil {
			log.Println(err)
		} else {
			res4, err := collector.OneType(syscall.AF_INET)
			if err != nil {
				log.Println(err)
			}
			record := collector.NetlinkResult{Time: time.Now(), IPv6: res6, IPv4: res4}
			b, err := json.Marshal(record)
			if err != nil {
				log.Println(err)
			}
			o.WriteString(string(b))
			o.WriteString("\n")
		}
	}
}