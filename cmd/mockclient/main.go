package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yedamao/encoding"
	"github.com/yedamao/go_sgip/sgip/protocol"
	"github.com/yedamao/go_sgip/sgip/sgiptest"
)

var (
	host   = flag.String("host", "localhost", "SP receiver host")
	port   = flag.Int("port", 8001, "SP receiver port")
	name   = flag.String("name", "", "Login Name")
	passwd = flag.String("passwd", "", "Login Password")

	spNumber   = flag.String("sp-number", "", "SP的接入号码")
	userNumber = flag.String("user-number", "", "发送短消息的用户手机号，手机号码前加“86”国别标志")
	msg        = flag.String("msg", "", "短信内容")
	report     = flag.String("report", "", "短信报告，格式: 状态码|seq1|seq2|seq3，如: 0|xxxx|xxxx|xxx")

	sleep = flag.Int("sleep", 1, "sleep some seconds after receive Deliver response")
)

func init() {
	flag.Parse()
}

func parseMsgSeq(msgId string) [3]uint32 {
	splits := strings.SplitN(msgId, "|", 3)
	seq1, _ := strconv.ParseInt(splits[0], 10, 32)
	seq2, _ := strconv.ParseInt(splits[1], 10, 64)
	seq3, _ := strconv.ParseInt(splits[2], 10, 64)
	return [3]uint32{uint32(seq1), uint32(seq2), uint32(seq3)}
}

func main() {

	client, err := sgiptest.NewSMGClient(*host, *port, *name, *passwd)
	if err != nil {
		fmt.Println("Connection Err:", err)
		return
	}

	fmt.Println("connect succ")

	// encoding msg
	content := encoding.UTF82GBK([]byte(*msg))

	if len(content) > 140 {
		fmt.Println("msg Err: not suport long sms")
	}

	fmt.Println("----- Deliver single msg -----")
	// Send sms
	if report != nil {
		split := strings.SplitN(*report, "|", 2)
		state, _ := strconv.ParseInt(split[0], 10, 32)
		msgSeq := parseMsgSeq(split[1])
		err = client.Report(msgSeq, 0, *userNumber, int(state), 0)
		if err != nil {
			fmt.Println("Report: ", err)
		}
	}
	if msg != nil {
		err = client.Deliver(*userNumber, *spNumber, 0, 0, protocol.GBK, content)
		if err != nil {
			fmt.Println("Deliver: ", err)
		}
	}

	for {
		op, err := client.Read() // This is blocking
		if err != nil {
			fmt.Println("Read Err:", err)
			break
		}

		fmt.Println(op)

		switch op.GetHeader().CmdId {
		case protocol.SGIP_DELIVER_REP:
			time.Sleep(time.Duration(*sleep) * time.Second)
			client.Unbind()
		case protocol.SGIP_REPORT_REP:
			time.Sleep(time.Duration(*sleep) * time.Second)
			client.Unbind()
		case protocol.SGIP_UNBIND_REP:
			fmt.Println("unbind response")
			break
		default:
			fmt.Printf("Unexpect CmdId: %0x\n", op.GetHeader().CmdId)
			fmt.Println("MSG ID:", op.GetHeader().Sequence)
		}
	}

	fmt.Println("ending...")
}
