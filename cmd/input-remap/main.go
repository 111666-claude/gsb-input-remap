// Command input-remap 跑按键重映射样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/remap"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("input-remap", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "conflict", "conflict / rebind / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample == "conflict" {
		// 同一个键两条绑定：优先级 1 的是 jump，优先级 2 的是 attack。
		keymap := remap.NewMap([]remap.Binding{
			{Key: "k", Action: "jump", Priority: 1},
			{Key: "k", Action: "attack", Priority: 2},
		})
		fmt.Fprintf(stdout, "action=%s\n", keymap.Action("k"))
		return 0
	}
	if *sample == "rebind" {
		keymap := remap.NewMap([]remap.Binding{{Key: "k", Action: "jump", Priority: 1}})
		ok := keymap.Rebind("k", "", 1)
		fmt.Fprintf(stdout, "ok=%v action=%s\n", ok, keymap.Action("k"))
		return 0
	}
	if *sample == "scan" {
		bindings := make([]remap.Binding, 0, 3000)
		for index := 0; index < 3000; index++ {
			bindings = append(bindings, remap.Binding{Key: fmt.Sprintf("k-%d", index), Action: "jump", Priority: index})
		}
		keymap := remap.NewMap(bindings)
		for query := 0; query < 3000; query++ {
			keymap.Action("k-2999")
		}
		fmt.Fprintf(stdout, "scanned=%d\n", keymap.Scanned())
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample conflict|rebind|scan")
	return 2
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
