package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"runtime"

	"github.com/eiannone/keyboard"
	"github.com/gosuri/uilive"
)

var text string = ""

func clearConsole() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func selectOptions(options []string, writer *uilive.Writer) string {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	var printOptions func([]string, int)
	selectedOption := 0

	printOptions = func(options []string, selectedOption int) {
		fmt.Fprintf(writer.Newline(), text)
		// Print the list of options
		for i, option := range options {
			if i == selectedOption {
				fmt.Fprintln(writer.Newline(), fmt.Sprintf("# %s ", strings.Split(option, ":")[0]))
				//color.New(color.Bold, color.FgGreen).Fprintln(writer.Newline(), "# %s\n ", strings.Split(option, ":")[0])
			} else {
				fmt.Fprintln(writer.Newline(), fmt.Sprintf("  %s", strings.Split(option, ":")[0]))
			}
		}
		writer.Flush()
	}

	// Print the list of options
	printOptions(options, selectedOption)

	// Wait for user input
	for {
		_, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println(err)
			return "ERROR"
		}

		// Update the selected option based on user input
		switch {
		case key == keyboard.KeyArrowUp:
			if selectedOption > 0 {
				selectedOption--
			} else if selectedOption <= 0 {
				selectedOption = len(options) - 1
			}
			printOptions(options, selectedOption)
		case key == keyboard.KeyArrowDown:
			if selectedOption < len(options)-1 {
				selectedOption++
			} else if selectedOption > len(options)-2 {
				selectedOption = 0
			}
			printOptions(options, selectedOption)
		case key == 13: // Enter key
			return options[selectedOption]
		}
	}

}

func FileToLines(filePath string) (lines []string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	return
}

func main() {
	// load input file
	lines, err := FileToLines("application.txt")
	if err != nil {
		fmt.Print("ERROR")
		return
	}

	lineIdx := 0
	var entryIndices = map[string]int{}
	var entryPoint string = ""

	var writer = uilive.New()
	writer.RefreshInterval = time.Hour

	for {
		if lineIdx >= len(lines) {
			fmt.Print("ERROR: lineIdx too big")
			return
		}

		// hide cursor
		fmt.Printf("\033[?25l")
		// show cursor
		// fmt.Printf("\033[?25h")

		sIdx := strings.Index(lines[lineIdx], "#")
		// is head of a text block
		if sIdx >= 0 {
			entryPoint = lines[lineIdx]
			entryIndices[entryPoint] = lineIdx
			lineIdx++
		}
		// when option found, print the text
		// check if is a option ( when the first option appears, it is the end of the above text)

		var options []string

		for {
			// step through text input to check for options after a discription text.
			// after no option is found anymore, run the selectOptions function to await input from user.

			if strings.Index(lines[lineIdx], ">> ") >= 0 {
				option := strings.Trim(lines[lineIdx], ">> ")
				options = append(options, option)

				lineIdx++
			} else if options != nil {
				writer.Start()
				selected := selectOptions(options, writer)
				writer.Stop()
				clearConsole()
				// check what the selected option referces to (text or file)
				slice := strings.Split(selected, ":")
				if slice[1] == "file" {
					// open a file
					// TODO: check on linux
					fmt.Println("start! ")
					cmd := exec.Command("cmd", "/C start "+slice[2])
					err := cmd.Start()
					if err != nil {
						fmt.Println("ERROR")
					}
					err = cmd.Wait()
					if err != nil {
						fmt.Println("ERROR")
					}
				} else if slice[1] == "exit" {
					// execute a command
					return
				} else {
					entryPoint = "#" + slice[1]
				}

				if val, ok := entryIndices[entryPoint]; ok {
					lineIdx = val + 1
				} else {
					for i := range lines {
						if strings.Contains(lines[i], entryPoint) {
							lineIdx = i + 1
							entryIndices[entryPoint] = i
							//fmt.Print("found! ")
							break
						}
					}
				}
				lineIdx = entryIndices[entryPoint] + 1
				text = ""
				break
			} else {
				break
			}
		}

		text = text + lines[lineIdx] + "\n"
		lineIdx++
	}
}
