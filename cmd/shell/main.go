// File: cmd/shell/main.go
// - main CLI binary for ChooChoo

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"net/http"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("ChooChoo CLI - Type 'exit' to quit")
	for {
		fmt.Print("choo> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "exit" {
			break
		}
		runQuery(line)
	}
}

func runQuery(input string) {
	payload := map[string]string{"q": input}
	data, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8080/query", "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		fmt.Println("Decode error:", err)
		return
	}
	printJSON(out)
}

func printJSON(obj map[string]interface{}) {
	for k, v := range obj {
		fmt.Printf("%s: %v\n", k, v)
	}
}
